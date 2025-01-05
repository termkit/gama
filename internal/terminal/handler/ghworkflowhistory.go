package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termkit/gama/internal/config"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/gama/pkg/browser"
	"github.com/termkit/skeleton"
)

// -----------------------------------------------------------------------------
// Model Definition
// -----------------------------------------------------------------------------

type ModelGithubWorkflowHistory struct {
	// Core dependencies
	skeleton *skeleton.Skeleton
	github   gu.UseCase

	// UI Components
	Help                 help.Model
	keys                 githubWorkflowHistoryKeyMap
	tableWorkflowHistory table.Model
	status               *ModelStatus
	modelTabOptions      *ModelTabOptions

	// Table state
	tableReady     bool
	tableStyle     lipgloss.Style
	updateRound    int
	workflows      []gu.Workflow
	lastRepository string

	// Live mode state
	liveMode         bool
	liveModeInterval time.Duration

	// Workflow state
	selectedWorkflowID int64

	// Context management
	syncWorkflowHistoryContext context.Context
	cancelSyncWorkflowHistory  context.CancelFunc

	// Shared state
	selectedRepository *SelectedRepository
}

type workflowHistoryUpdateMsg struct {
	UpdateAfter time.Duration
}

// -----------------------------------------------------------------------------
// Constructor & Initialization
// -----------------------------------------------------------------------------

func SetupModelGithubWorkflowHistory(sk *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubWorkflowHistory {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	m := &ModelGithubWorkflowHistory{
		// Initialize core dependencies
		skeleton: sk,
		github:   githubUseCase,

		// Initialize UI components
		Help:            help.New(),
		keys:            githubWorkflowHistoryKeys,
		status:          SetupModelStatus(sk),
		modelTabOptions: NewOptions(sk, SetupModelStatus(sk)),

		// Initialize state
		selectedRepository:         NewSelectedRepository(),
		syncWorkflowHistoryContext: context.Background(),
		cancelSyncWorkflowHistory:  func() {},
		liveMode:                   cfg.Settings.LiveMode.Enabled,
		liveModeInterval:           cfg.Settings.LiveMode.Interval,
		tableStyle:                 setupTableStyle(),
	}

	// Setup table
	m.tableWorkflowHistory = setupWorkflowHistoryTable()

	return m
}

func setupTableStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		MarginLeft(1)
}

func setupWorkflowHistoryTable() table.Model {
	t := table.New(
		table.WithColumns(tableColumnsWorkflowHistory),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Apply styles
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
	t.SetStyles(s)

	t.KeyMap = table.KeyMap{
		LineUp:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		LineDown:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		PageUp:     key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up")),
		PageDown:   key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "page down")),
		GotoTop:    key.NewBinding(key.WithKeys("home"), key.WithHelp("home", "go to start")),
		GotoBottom: key.NewBinding(key.WithKeys("end"), key.WithHelp("end", "go to end")),
	}

	return t
}

// -----------------------------------------------------------------------------
// Bubbletea Model Implementation
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) Init() tea.Cmd {
	m.setupOptions()
	m.startLiveMode()
	return tea.Batch(m.modelTabOptions.Init())
}

func (m *ModelGithubWorkflowHistory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle repository changes
	if cmd := m.handleRepositoryChange(); cmd != nil {
		return m, cmd
	}

	cursor := m.tableWorkflowHistory.Cursor()
	if m.workflows != nil && cursor >= 0 && cursor < len(m.workflows) {
		m.selectedWorkflowID = m.workflows[cursor].ID
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle different message types
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKeyMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	case workflowHistoryUpdateMsg:
		cmd = m.handleUpdateMsg(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update UI components
	if cmd = m.updateUIComponents(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubWorkflowHistory) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		m.renderTable(),
		m.modelTabOptions.View(),
		m.status.View(),
		m.renderHelp(),
	)
}

// -----------------------------------------------------------------------------
// Event Handlers
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keys.Refresh):
		go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
		return nil
	case key.Matches(msg, m.keys.LiveMode):
		return m.toggleLiveMode()
	}
	return nil
}

func (m *ModelGithubWorkflowHistory) handleUpdateMsg(msg workflowHistoryUpdateMsg) tea.Cmd {
	go func() {
		time.Sleep(msg.UpdateAfter)
		m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
		m.skeleton.TriggerUpdate()
	}()
	return nil
}

// -----------------------------------------------------------------------------
// Live Mode Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) startLiveMode() {
	go func() {
		for range time.NewTicker(m.liveModeInterval).C {
			if m.liveMode {
				m.skeleton.TriggerUpdateWithMsg(workflowHistoryUpdateMsg{UpdateAfter: time.Nanosecond})
			}
		}
	}()
}

func (m *ModelGithubWorkflowHistory) toggleLiveMode() tea.Cmd {
	m.liveMode = !m.liveMode
	if m.liveMode {
		m.status.SetSuccessMessage("Live mode enabled")
		m.skeleton.UpdateWidgetValue("live", "Live Mode: On")
	} else {
		m.status.SetSuccessMessage("Live mode disabled")
		m.skeleton.UpdateWidgetValue("live", "Live Mode: Off")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Repository Change Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) handleRepositoryChange() tea.Cmd {
	if m.lastRepository == m.selectedRepository.RepositoryName {
		return nil
	}

	if m.cancelSyncWorkflowHistory != nil {
		m.cancelSyncWorkflowHistory()
	}

	m.lastRepository = m.selectedRepository.RepositoryName
	m.syncWorkflowHistoryContext, m.cancelSyncWorkflowHistory = context.WithCancel(context.Background())

	go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
	return nil
}

// -----------------------------------------------------------------------------
// Workflow History Sync
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) syncWorkflowHistory(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.initializeSyncState()
	workflowHistory, err := m.fetchWorkflowHistory(ctx)
	if err != nil {
		return
	}

	m.processWorkflowHistory(workflowHistory)
}

func (m *ModelGithubWorkflowHistory) initializeSyncState() {
	m.tableReady = false
	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s@%s] Fetching workflow history...",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
	m.modelTabOptions.SetStatus(StatusWait)
	m.clearWorkflowHistory()
}

func (m *ModelGithubWorkflowHistory) clearWorkflowHistory() {
	m.tableWorkflowHistory.SetRows([]table.Row{})
	m.workflows = nil
}

func (m *ModelGithubWorkflowHistory) fetchWorkflowHistory(ctx context.Context) (*gu.GetWorkflowHistoryOutput, error) {
	history, err := m.github.GetWorkflowHistory(ctx, gu.GetWorkflowHistoryInput{
		Repository: m.selectedRepository.RepositoryName,
		Branch:     m.selectedRepository.BranchName,
	})

	if err != nil {
		if !errors.Is(err, context.Canceled) {
			m.status.SetError(err)
			m.status.SetErrorMessage("Workflow history cannot be listed")
		}
		return nil, err
	}

	return history, nil
}

// -----------------------------------------------------------------------------
// Workflow History Processing
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) processWorkflowHistory(history *gu.GetWorkflowHistoryOutput) {
	if len(history.Workflows) == 0 {
		m.handleEmptyWorkflowHistory()
		return
	}

	m.workflows = history.Workflows
	m.updateWorkflowTable()
	m.finalizeUpdate()
}

func (m *ModelGithubWorkflowHistory) handleEmptyWorkflowHistory() {
	m.modelTabOptions.SetStatus(StatusNone)
	m.status.SetDefaultMessage(fmt.Sprintf("[%s@%s] No workflow history found.",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
}

func (m *ModelGithubWorkflowHistory) updateWorkflowTable() {
	rows := make([]table.Row, 0, len(m.workflows))
	for _, workflow := range m.workflows {
		rows = append(rows, table.Row{
			workflow.WorkflowName,
			workflow.ActionName,
			workflow.TriggeredBy,
			workflow.StartedAt,
			workflow.Status,
			workflow.Duration,
		})
	}
	m.tableWorkflowHistory.SetRows(rows)
}

func (m *ModelGithubWorkflowHistory) finalizeUpdate() {
	m.tableReady = true
	m.tableWorkflowHistory.SetCursor(0)
	m.modelTabOptions.SetStatus(StatusIdle)
	m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow history fetched.",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
}

// -----------------------------------------------------------------------------
// UI Component Updates
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) updateUIComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update table and handle navigation
	m.tableWorkflowHistory, cmd = m.tableWorkflowHistory.Update(msg)
	cmds = append(cmds, cmd)

	// Update tab options
	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// -----------------------------------------------------------------------------
// UI Rendering
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) renderTable() string {
	m.updateTableDimensions()
	return m.tableStyle.Render(m.tableWorkflowHistory.View())
}

func (m *ModelGithubWorkflowHistory) renderHelp() string {
	helpStyle := WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)
	return helpStyle.Render(m.ViewHelp())
}

func (m *ModelGithubWorkflowHistory) updateTableDimensions() {
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
}

// -----------------------------------------------------------------------------
// Option Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflowHistory) setupOptions() {
	m.modelTabOptions.AddOption("Open in browser", m.openInBrowser)
	m.modelTabOptions.AddOption("Rerun failed jobs", m.rerunFailedJobs)
	m.modelTabOptions.AddOption("Rerun workflow", m.rerunWorkflow)
	m.modelTabOptions.AddOption("Cancel workflow", m.cancelWorkflow)
}

func (m *ModelGithubWorkflowHistory) openInBrowser() {
	m.status.SetProgressMessage("Opening in browser...")

	url := fmt.Sprintf("https://github.com/%s/actions/runs/%d",
		m.selectedRepository.RepositoryName,
		m.selectedWorkflowID)

	if err := browser.OpenInBrowser(url); err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Failed to open in browser")
		return
	}
	m.status.SetSuccessMessage("Opened in browser")
}

func (m *ModelGithubWorkflowHistory) rerunFailedJobs() {
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

func (m *ModelGithubWorkflowHistory) rerunWorkflow() {
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

func (m *ModelGithubWorkflowHistory) cancelWorkflow() {
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
