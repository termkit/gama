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
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
)

type ModelGithubWorkflowHistory struct {
	githubUseCase gu.UseCase

	Help                 help.Model
	Keys                 keyMap
	Viewport             *viewport.Model
	tableWorkflowHistory table.Model
	modelError           hdlerror.ModelError

	SelectedRepository *hdltypes.SelectedRepository
	updateRound        int

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

	return &ModelGithubWorkflowHistory{
		Help:                 help.New(),
		Keys:                 keys,
		githubUseCase:        githubUseCase,
		tableWorkflowHistory: tableWorkflowHistory,
		modelError:           hdlerror.SetupModelError(),
		SelectedRepository:   selectedRepository,
	}
}

func (m *ModelGithubWorkflowHistory) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubWorkflowHistory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.lastRepository != m.SelectedRepository.RepositoryName {
		go m.updateWorkflowHistory()
		m.lastRepository = m.SelectedRepository.RepositoryName
	}
	var cmd tea.Cmd
	//switch msg := msg.(type) {
	//case tea.KeyMsg:
	//	switch msg.String() {
	//	case "q", "ctrl+c":
	//		return m, tea.Quit
	//	}
	//}
	m.tableWorkflowHistory, cmd = m.tableWorkflowHistory.Update(msg)
	return m, cmd
}

func (m *ModelGithubWorkflowHistory) updateWorkflowHistory() {
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s] Fetching workflow history...", m.SelectedRepository.RepositoryName))

	// delete all rows
	m.tableWorkflowHistory.SetRows([]table.Row{})

	workflowHistory, err := m.githubUseCase.GetWorkflowHistory(context.Background(), gu.GetWorkflowHistoryInput{
		Repository: m.SelectedRepository.RepositoryName,
	})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow history cannot be listed")
	}

	if len(workflowHistory.Workflows) == 0 {
		m.modelError.SetDefaultMessage(fmt.Sprintf("[%s] No workflows found.", m.SelectedRepository.RepositoryName))
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
	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s] Workflow history fetched.", m.SelectedRepository.RepositoryName))
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

	m.tableWorkflowHistory.SetHeight(termHeight - 16)

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableWorkflowHistory.View()))

	return doc.String()
}

func (m *ModelGithubWorkflowHistory) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubWorkflowHistory) IsError() bool {
	return m.modelError.IsError()
}
