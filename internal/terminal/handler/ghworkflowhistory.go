package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/termkit/gama/internal/config"
	"github.com/termkit/skeleton"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/gama/internal/terminal/handler/status"
	"github.com/termkit/gama/internal/terminal/handler/taboptions"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/gama/pkg/browser"
)

type ModelGithubWorkflowHistory struct {
	skeleton *skeleton.Skeleton
	// current handler's properties
	tableReady                 bool
	liveMode                   bool
	liveModeInterval           time.Duration
	tableStyle                 lipgloss.Style
	updateRound                int
	selectedWorkflowID         int64
	lastRepository             string
	syncWorkflowHistoryContext context.Context
	cancelSyncWorkflowHistory  context.CancelFunc
	workflows                  []gu.Workflow

	// shared properties
	selectedRepository *hdltypes.SelectedRepository

	// use cases
	github gu.UseCase

	// keymap
	keys githubWorkflowHistoryKeyMap

	// models
	Help                 help.Model
	tableWorkflowHistory table.Model
	status               *status.ModelStatus

	modelTabOptions *taboptions.Options
}

type workflowHistoryUpdateMsg struct {
	UpdateAfter time.Duration
}

func SetupModelGithubWorkflowHistory(sk *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubWorkflowHistory {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

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

	var tableStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).MarginLeft(1)

	modelStatus := status.SetupModelStatus(sk)
	tabOptions := taboptions.NewOptions(sk, &modelStatus)

	return &ModelGithubWorkflowHistory{
		skeleton:                   sk,
		liveMode:                   cfg.Settings.LiveMode.Enabled,
		liveModeInterval:           cfg.Settings.LiveMode.Interval,
		Help:                       help.New(),
		keys:                       githubWorkflowHistoryKeys,
		github:                     githubUseCase,
		tableWorkflowHistory:       tableWorkflowHistory,
		status:                     &modelStatus,
		selectedRepository:         hdltypes.NewSelectedRepository(),
		modelTabOptions:            tabOptions,
		syncWorkflowHistoryContext: context.Background(),
		cancelSyncWorkflowHistory:  func() {},
		tableStyle:                 tableStyle,
	}
}

func (m *ModelGithubWorkflowHistory) Init() tea.Cmd {
	m.setupOptions()
	m.ToggleLiveMode()
	return tea.Batch(m.modelTabOptions.Init())
}
func (m *ModelGithubWorkflowHistory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.lastRepository != m.selectedRepository.RepositoryName {
		m.tableReady = false
		m.cancelSyncWorkflowHistory() // cancel previous sync

		m.lastRepository = m.selectedRepository.RepositoryName

		m.syncWorkflowHistoryContext, m.cancelSyncWorkflowHistory = context.WithCancel(context.Background())
		go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
	}

	if m.workflows != nil {
		m.selectedWorkflowID = m.workflows[m.tableWorkflowHistory.Cursor()].ID
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Refresh):
			go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
		case key.Matches(msg, m.keys.LiveMode):
			m.liveMode = !m.liveMode
			if m.liveMode {
				m.status.SetSuccessMessage("Live mode enabled")
				m.skeleton.UpdateWidgetValue("live", "Live Mode: On")
			} else {
				m.status.SetSuccessMessage("Live mode disabled")
				m.skeleton.UpdateWidgetValue("live", "Live Mode: Off")
			}
		}
	case workflowHistoryUpdateMsg:
		go func() {
			time.Sleep(msg.UpdateAfter)
			m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
			m.skeleton.TriggerUpdate()
		}()
	}

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	m.tableWorkflowHistory, cmd = m.tableWorkflowHistory.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubWorkflowHistory) View() string {
	helpWindowStyle := hdltypes.WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)

	termWidth := m.skeleton.GetTerminalWidth()
	termHeight := m.skeleton.GetTerminalHeight()

	var tableWidth int
	for _, t := range tableColumnsWorkflowHistory {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsWorkflowHistory
	widthDiff := termWidth - tableWidth

	if widthDiff > 0 {
		if m.updateRound%2 == 0 {
			newTableColumns[0].Width += widthDiff - 18
		} else {
			newTableColumns[1].Width += widthDiff - 18
		}
		m.updateRound++
		m.tableWorkflowHistory.SetColumns(newTableColumns)
	}

	m.tableWorkflowHistory.SetHeight(termHeight - 17)

	return lipgloss.JoinVertical(lipgloss.Top,
		m.tableStyle.Render(m.tableWorkflowHistory.View()), m.modelTabOptions.View(),
		m.status.View(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelGithubWorkflowHistory) setupOptions() {
	openInBrowser := func() {
		m.status.SetProgressMessage("Opening in browser...")

		var selectedWorkflow = fmt.Sprintf("https://github.com/%s/actions/runs/%d", m.selectedRepository.RepositoryName, m.selectedWorkflowID)

		err := browser.OpenInBrowser(selectedWorkflow)
		if err != nil {
			m.status.SetError(err)
			m.status.SetErrorMessage("Failed to open in browser")
			return
		}
		m.status.SetSuccessMessage("Opened in browser")
	}

	reRunFailedJobs := func() {
		m.status.SetProgressMessage("Re-running failed jobs...")

		_, err := m.github.ReRunFailedJobs(context.Background(), gu.ReRunFailedJobsInput{
			Repository: m.selectedRepository.RepositoryName,
			WorkflowID: m.selectedWorkflowID,
		})

		if err != nil {
			m.status.SetError(err)
			m.status.SetErrorMessage("Failed to re-run failed jobs")
			return
		}

		m.status.SetSuccessMessage("Re-ran failed jobs")
	}

	reRunWorkflow := func() {
		m.status.SetProgressMessage("Re-running workflow...")

		_, err := m.github.ReRunWorkflow(context.Background(), gu.ReRunWorkflowInput{
			Repository: m.selectedRepository.RepositoryName,
			WorkflowID: m.selectedWorkflowID,
		})

		if err != nil {
			m.status.SetError(err)
			m.status.SetErrorMessage("Failed to re-run workflow")
			return
		}

		m.status.SetSuccessMessage("Re-ran workflow")
	}

	cancelWorkflow := func() {
		m.status.SetProgressMessage("Canceling workflow...")

		_, err := m.github.CancelWorkflow(context.Background(), gu.CancelWorkflowInput{
			Repository: m.selectedRepository.RepositoryName,
			WorkflowID: m.selectedWorkflowID,
		})

		if err != nil {
			m.status.SetError(err)
			m.status.SetErrorMessage("Failed to cancel workflow")
			return
		}

		m.status.SetSuccessMessage("Canceled workflow")
	}
	m.modelTabOptions.AddOption("Open in browser", openInBrowser)
	m.modelTabOptions.AddOption("Rerun failed jobs", reRunFailedJobs)
	m.modelTabOptions.AddOption("Rerun workflow", reRunWorkflow)
	m.modelTabOptions.AddOption("Cancel workflow", cancelWorkflow)
}

func (m *ModelGithubWorkflowHistory) ToggleLiveMode() {
	// send UpdateWorkflowHistoryMsg to update the workflow history every 5 seconds with ticker
	// send only if liveMode is true
	go func() {
		t := time.NewTicker(m.liveModeInterval)
		for {
			select {
			case <-t.C:
				if m.liveMode {
					m.skeleton.TriggerUpdateWithMsg(workflowHistoryUpdateMsg{UpdateAfter: time.Nanosecond})
				}
			}
		}
	}()
}

func (m *ModelGithubWorkflowHistory) syncWorkflowHistory(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.tableReady = false
	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s@%s] Fetching workflow history...",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
	m.modelTabOptions.SetStatus(taboptions.OptionWait)

	// delete all rows
	m.tableWorkflowHistory.SetRows([]table.Row{})

	// delete old workflows
	m.workflows = nil

	workflowHistory, err := m.github.GetWorkflowHistory(ctx, gu.GetWorkflowHistoryInput{
		Repository: m.selectedRepository.RepositoryName,
		Branch:     m.selectedRepository.BranchName,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Workflow history cannot be listed")
		return
	}

	if len(workflowHistory.Workflows) == 0 {
		m.modelTabOptions.SetStatus(taboptions.OptionNone)
		m.status.SetDefaultMessage(fmt.Sprintf("[%s@%s] No workflows found.", m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
		return
	}

	m.workflows = workflowHistory.Workflows

	var tableRowsWorkflowHistory []table.Row
	for _, workflowRun := range m.workflows {
		tableRowsWorkflowHistory = append(tableRowsWorkflowHistory, table.Row{
			workflowRun.WorkflowName,
			workflowRun.ActionName,
			workflowRun.TriggeredBy,
			workflowRun.StartedAt,
			workflowRun.Status,
			workflowRun.Duration,
		})
	}

	m.tableReady = true
	m.tableWorkflowHistory.SetRows(tableRowsWorkflowHistory)
	m.tableWorkflowHistory.SetCursor(0)
	m.modelTabOptions.SetStatus(taboptions.OptionIdle)
	m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow history fetched.", m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
}
