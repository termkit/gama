package ghworkflowhistory

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	"github.com/termkit/gama/internal/terminal/handler/taboptions"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
)

type ModelGithubWorkflowHistory struct {
	githubUseCase gu.UseCase

	Help                 help.Model
	Keys                 keyMap
	Viewport             *viewport.Model
	tableWorkflowHistory table.Model
	modelError           hdlerror.ModelError

	modelTabOptions       tea.Model
	actualModelTabOptions *taboptions.Options

	SelectedRepository *hdltypes.SelectedRepository
	updateRound        int

	isTableFocused bool

	lastRepository string
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func SetupModelGithubWorkflowHistory(githubUseCase gu.UseCase, selectedRepository *hdltypes.SelectedRepository) *ModelGithubWorkflowHistory {
	tableRowsWorkflowHistory := []table.Row{}

	tableWorkflowHistory := table.New(
		table.WithColumns(tableColumnsWorkflowHistory),
		table.WithRows(tableRowsWorkflowHistory),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tableWorkflowHistory.SetStyles(s)

	tabOptions := taboptions.NewOptions()

	return &ModelGithubWorkflowHistory{
		Help:                  help.New(),
		Keys:                  keys,
		githubUseCase:         githubUseCase,
		tableWorkflowHistory:  tableWorkflowHistory,
		modelError:            hdlerror.SetupModelError(),
		SelectedRepository:    selectedRepository,
		modelTabOptions:       tabOptions,
		actualModelTabOptions: tabOptions,
	}
}

func (m *ModelGithubWorkflowHistory) Init() tea.Cmd {
	m.actualModelTabOptions.AddOption("Open in browser")
	m.actualModelTabOptions.AddOption("Rerun failed jobs")
	m.actualModelTabOptions.AddOption("Rerun workflow")
	m.actualModelTabOptions.AddOption("Cancel workflow")

	return nil
}

func (m *ModelGithubWorkflowHistory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.lastRepository != m.SelectedRepository.RepositoryName {
		go m.updateWorkflowHistory()
		m.lastRepository = m.SelectedRepository.RepositoryName
	}
	//var cmd tea.Cmd

	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.isTableFocused {
				m.isTableFocused = false
				m.tableWorkflowHistory.Blur()
			} else {
				m.isTableFocused = true
				m.tableWorkflowHistory.Focus()

			}
		}
	}

	if !m.isTableFocused {
		//m.modelTabOptions.Update(msg)
		m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
		cmds = append(cmds, cmd)
	}

	//m.tableWorkflowHistory, cmd := m.tableWorkflowHistory.Update(msg)
	m.tableWorkflowHistory, cmd = m.tableWorkflowHistory.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *ModelGithubWorkflowHistory) updateWorkflowHistory() {
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching workflow history...", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))

	// delete all rows
	m.tableWorkflowHistory.SetRows([]table.Row{})

	workflowHistory, err := m.githubUseCase.GetWorkflowHistory(context.Background(), gu.GetWorkflowHistoryInput{
		Repository: m.SelectedRepository.RepositoryName,
		Branch:     m.SelectedRepository.BranchName,
	})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow history cannot be listed")
	}

	if len(workflowHistory.Workflows) == 0 {
		m.modelError.SetDefaultMessage(fmt.Sprintf("[%s@%s] No workflows found.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
		return
	}

	var tableRowsWorkflowHistory []table.Row
	for _, workflowRun := range workflowHistory.Workflows {
		tableRowsWorkflowHistory = append(tableRowsWorkflowHistory, table.Row{
			workflowRun.WorkflowName,
			workflowRun.ActionName,
			workflowRun.TriggeredBy,
			workflowRun.StartedAt,
			workflowRun.Conslusion,
			workflowRun.Duration,
		})
	}

	m.tableWorkflowHistory.SetRows(tableRowsWorkflowHistory)
	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow history fetched.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
}

func (m *ModelGithubWorkflowHistory) View() string {
	termWidth := m.Viewport.Width
	termHeight := m.Viewport.Height

	var tableWidth int
	for _, t := range tableColumnsWorkflowHistory {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsWorkflowHistory
	widthDiff := termWidth - tableWidth

	if widthDiff > 0 {
		if m.updateRound%2 == 0 {
			newTableColumns[0].Width += widthDiff - 19
		} else {
			newTableColumns[1].Width += widthDiff - 19
		}
		m.updateRound++
		m.tableWorkflowHistory.SetColumns(newTableColumns)
	}

	m.tableWorkflowHistory.SetHeight(termHeight - 17)

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableWorkflowHistory.View()))

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.actualModelTabOptions.View())
}

func (m *ModelGithubWorkflowHistory) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubWorkflowHistory) IsError() bool {
	return m.modelError.IsError()
}
