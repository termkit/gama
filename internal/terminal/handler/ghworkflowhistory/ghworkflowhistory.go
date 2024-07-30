package ghworkflowhistory

import (
	"context"
	"errors"
	"fmt"
	ts "github.com/termkit/gama/internal/terminal/style"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	"github.com/termkit/gama/internal/terminal/handler/taboptions"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/gama/pkg/browser"
)

type ModelGithubWorkflowHistory struct {
	// current handler's properties
	tableReady                 bool
	updateRound                int
	selectedWorkflowID         int64
	isTableFocused             bool
	lastRepository             string
	syncWorkflowHistoryContext context.Context
	cancelSyncWorkflowHistory  context.CancelFunc
	Workflows                  []gu.Workflow

	// shared properties
	SelectedRepository *hdltypes.SelectedRepository

	// use cases
	github gu.UseCase

	// keymap
	Keys keyMap

	// models
	Help                 help.Model
	Viewport             *viewport.Model
	tableWorkflowHistory table.Model
	modelError           *hdlerror.ModelError

	modelTabOptions *taboptions.Options
}

func SetupModelGithubWorkflowHistory(githubUseCase gu.UseCase) *ModelGithubWorkflowHistory {
	var tableRowsWorkflowHistory []table.Row

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

	modelError := hdlerror.SetupModelError()
	tabOptions := taboptions.NewOptions(&modelError)

	return &ModelGithubWorkflowHistory{
		Viewport:                   hdltypes.NewTerminalViewport(),
		Help:                       help.New(),
		Keys:                       keys,
		github:                     githubUseCase,
		tableWorkflowHistory:       tableWorkflowHistory,
		modelError:                 &modelError,
		SelectedRepository:         hdltypes.NewSelectedRepository(),
		modelTabOptions:            tabOptions,
		syncWorkflowHistoryContext: context.Background(),
		cancelSyncWorkflowHistory:  func() {},
	}
}

type UpdateWorkflowHistoryMsg struct {
	UpdateAfter time.Duration
}

func (m *ModelGithubWorkflowHistory) Init() tea.Cmd {
	m.setupOptions()
	return tea.Batch(
		m.modelTabOptions.Init(),
		func() tea.Msg {
			return UpdateWorkflowHistoryMsg{UpdateAfter: time.Second * 1}
		},
		m.modelError.Init(),
	)
}

func (m *ModelGithubWorkflowHistory) setupOptions() {
	openInBrowser := func() {
		m.modelError.SetProgressMessage("Opening in browser...")

		var selectedWorkflow = fmt.Sprintf("https://github.com/%s/actions/runs/%d", m.SelectedRepository.RepositoryName, m.selectedWorkflowID)

		err := browser.OpenInBrowser(selectedWorkflow)
		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("Failed to open in browser")
			return
		}
		m.modelError.SetSuccessMessage("Opened in browser")
	}

	reRunFailedJobs := func() {
		m.modelError.SetProgressMessage("Re-running failed jobs...")

		_, err := m.github.ReRunFailedJobs(context.Background(), gu.ReRunFailedJobsInput{
			Repository: m.SelectedRepository.RepositoryName,
			WorkflowID: m.selectedWorkflowID,
		})

		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("Failed to re-run failed jobs")
			return
		}

		m.modelError.SetSuccessMessage("Re-ran failed jobs")
	}

	reRunWorkflow := func() {
		m.modelError.SetProgressMessage("Re-running workflow...")

		_, err := m.github.ReRunWorkflow(context.Background(), gu.ReRunWorkflowInput{
			Repository: m.SelectedRepository.RepositoryName,
			WorkflowID: m.selectedWorkflowID,
		})

		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("Failed to re-run workflow")
			return
		}

		m.modelError.SetSuccessMessage("Re-ran workflow")
	}

	cancelWorkflow := func() {
		m.modelError.SetProgressMessage("Canceling workflow...")

		_, err := m.github.CancelWorkflow(context.Background(), gu.CancelWorkflowInput{
			Repository: m.SelectedRepository.RepositoryName,
			WorkflowID: m.selectedWorkflowID,
		})

		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("Failed to cancel workflow")
			return
		}

		m.modelError.SetSuccessMessage("Canceled workflow")
	}
	m.modelTabOptions.AddOption("Open in browser", openInBrowser)
	m.modelTabOptions.AddOption("Rerun failed jobs", reRunFailedJobs)
	m.modelTabOptions.AddOption("Rerun workflow", reRunWorkflow)
	m.modelTabOptions.AddOption("Cancel workflow", cancelWorkflow)
}

func (m *ModelGithubWorkflowHistory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.lastRepository != m.SelectedRepository.RepositoryName {
		m.tableReady = false
		m.cancelSyncWorkflowHistory() // cancel previous sync

		m.lastRepository = m.SelectedRepository.RepositoryName

		m.syncWorkflowHistoryContext, m.cancelSyncWorkflowHistory = context.WithCancel(context.Background())
		go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
	}

	if m.Workflows != nil {
		m.selectedWorkflowID = m.Workflows[m.tableWorkflowHistory.Cursor()].ID
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Refresh):
			m.tableReady = false
			go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
		}
	case UpdateWorkflowHistoryMsg:
		go func() {
			time.Sleep(msg.UpdateAfter)
			m.tableReady = false
			m.syncWorkflowHistory(m.syncWorkflowHistoryContext) // TODO : may you use go routine here?
		}()
	}

	m.modelError, cmd = m.modelError.Update(msg)
	cmds = append(cmds, cmd)

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	m.tableWorkflowHistory, cmd = m.tableWorkflowHistory.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubWorkflowHistory) syncWorkflowHistory(ctx context.Context) {
	m.modelError.Reset()
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching workflow history...", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
	m.modelTabOptions.SetStatus(taboptions.OptionWait)

	// delete all rows
	m.tableWorkflowHistory.SetRows([]table.Row{})

	// delete old workflows
	m.Workflows = nil

	workflowHistory, err := m.github.GetWorkflowHistory(ctx, gu.GetWorkflowHistoryInput{
		Repository: m.SelectedRepository.RepositoryName,
		Branch:     m.SelectedRepository.BranchName,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow history cannot be listed")
		return
	}

	if len(workflowHistory.Workflows) == 0 {
		m.modelTabOptions.SetStatus(taboptions.OptionNone)
		m.modelError.SetDefaultMessage(fmt.Sprintf("[%s@%s] No workflows found.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
		return
	}

	m.Workflows = workflowHistory.Workflows

	var tableRowsWorkflowHistory []table.Row
	for _, workflowRun := range m.Workflows {
		tableRowsWorkflowHistory = append(tableRowsWorkflowHistory, table.Row{
			workflowRun.WorkflowName,
			workflowRun.ActionName,
			workflowRun.TriggeredBy,
			workflowRun.StartedAt,
			workflowRun.Conclusion,
			workflowRun.Duration,
		})
	}

	m.tableReady = true
	m.tableWorkflowHistory.SetRows(tableRowsWorkflowHistory)
	m.tableWorkflowHistory.SetCursor(0)
	m.modelTabOptions.SetStatus(taboptions.OptionIdle)
	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow history fetched.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
	go m.Update(m) // update model
}

func (m *ModelGithubWorkflowHistory) View() string {
	var baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).MarginLeft(1)
	helpWindowStyle := ts.WindowStyleHelp.Width(m.Viewport.Width - 4)

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
			newTableColumns[0].Width += widthDiff - 16
		} else {
			newTableColumns[1].Width += widthDiff - 16
		}
		m.updateRound++
		m.tableWorkflowHistory.SetColumns(newTableColumns)
	}

	m.tableWorkflowHistory.SetHeight(termHeight - 16)

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableWorkflowHistory.View()))

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.modelTabOptions.View(), m.ViewStatus(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelGithubWorkflowHistory) ViewStatus() string {
	return m.modelError.View()
}
