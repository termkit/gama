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

func SetupModelGithubWorkflowHistory(s *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubWorkflowHistory {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	modelStatus := SetupModelStatus(s)
	tabOptions := NewOptions(s, modelStatus)
	m := &ModelGithubWorkflowHistory{
		// Initialize core dependencies
		skeleton: s,
		github:   githubUseCase,

		// Initialize UI components
		Help:            help.New(),
		keys:            githubWorkflowHistoryKeys,
		status:          modelStatus,
		modelTabOptions: tabOptions,

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
	m.handleRepositoryChange()

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
		cmds = append(cmds, m.updateUIComponents(msg))
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
	ticker := time.NewTicker(m.liveModeInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if m.liveMode {
					m.skeleton.TriggerUpdateWithMsg(workflowHistoryUpdateMsg{
						UpdateAfter: time.Nanosecond,
					})
				}
			case <-m.syncWorkflowHistoryContext.Done():
				return
			}
		}
	}()
}

func (m *ModelGithubWorkflowHistory) toggleLiveMode() tea.Cmd {
	m.liveMode = !m.liveMode

	status := "Off"
	message := "Live mode disabled"

	if m.liveMode {
		status = "On"
		message = "Live mode enabled"
		// Trigger immediate update when enabling
		go m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
	}

	m.status.SetSuccessMessage(message)
	m.skeleton.UpdateWidgetValue("live", fmt.Sprintf("Live Mode: %s", status))

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

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	m.initializeSyncState()

	// Check context before proceeding
	if ctx.Err() != nil {
		m.status.SetDefaultMessage("Operation cancelled")
		return
	}

	workflowHistory, err := m.fetchWorkflowHistory(ctx)
	if err != nil {
		m.handleFetchError(err)
		return
	}

	m.processWorkflowHistory(workflowHistory)
}

func (m *ModelGithubWorkflowHistory) handleFetchError(err error) {
	switch {
	case errors.Is(err, context.Canceled):
		m.status.SetDefaultMessage("Workflow history fetch cancelled")
	case errors.Is(err, context.DeadlineExceeded):
		m.status.SetErrorMessage("Workflow history fetch timed out")
	default:
		m.status.SetError(err)
		m.status.SetErrorMessage(fmt.Sprintf("Failed to fetch workflow history: %v", err))
	}
}

func (m *ModelGithubWorkflowHistory) initializeSyncState() {
	m.tableReady = false
	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s] Fetching workflow history...", m.selectedRepository.RepositoryName))
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
	m.status.SetDefaultMessage(fmt.Sprintf("[%s] No workflow history found.",
		m.selectedRepository.RepositoryName))
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
	m.status.SetSuccessMessage(fmt.Sprintf("[%s] Workflow history fetched.", m.selectedRepository.RepositoryName))
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
	const (
		minTableWidth  = 80 // Minimum width to maintain readability
		tablePadding   = 18 // Account for borders and margins
		minColumnWidth = 10 // Minimum width for any column
	)

	termWidth := m.skeleton.GetTerminalWidth()
	termHeight := m.skeleton.GetTerminalHeight()

	if termWidth <= minTableWidth {
		return // Prevent table from becoming too narrow
	}

	var tableWidth int
	for _, t := range tableColumnsWorkflowHistory {
		tableWidth += t.Width
	}

	newTableColumns := make([]table.Column, len(tableColumnsWorkflowHistory))
	copy(newTableColumns, tableColumnsWorkflowHistory)

	widthDiff := termWidth - tableWidth - tablePadding
	if widthDiff > 0 {
		// Distribute extra width between workflow name and action name columns
		extraWidth := widthDiff / 2
		newTableColumns[0].Width = max(newTableColumns[0].Width+extraWidth, minColumnWidth)
		newTableColumns[1].Width = max(newTableColumns[1].Width+extraWidth, minColumnWidth)

		m.tableWorkflowHistory.SetColumns(newTableColumns)
	}

	// Ensure reasonable table height
	maxHeight := termHeight - 17
	if maxHeight > 0 {
		m.tableWorkflowHistory.SetHeight(maxHeight)
	}
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

	if err := m.github.ReRunFailedJobs(context.Background(), gu.ReRunFailedJobsInput{
		Repository: m.selectedRepository.RepositoryName,
		WorkflowID: m.selectedWorkflowID,
	}); err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Failed to re-run failed jobs")
		return
	}

	m.status.SetSuccessMessage("Re-ran failed jobs")
}

func (m *ModelGithubWorkflowHistory) rerunWorkflow() {
	if m.selectedWorkflowID == 0 {
		m.status.SetErrorMessage("No workflow selected")
		return
	}

	m.status.SetProgressMessage("Re-running workflow...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.github.ReRunWorkflow(ctx, gu.ReRunWorkflowInput{
		Repository: m.selectedRepository.RepositoryName,
		WorkflowID: m.selectedWorkflowID,
	}); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			m.status.SetErrorMessage("Workflow re-run request timed out")
		} else {
			m.status.SetError(err)
			m.status.SetErrorMessage(fmt.Sprintf("Failed to re-run workflow: %v", err))
		}
		return
	}

	m.status.SetSuccessMessage("Workflow re-run initiated")
	// Trigger refresh after short delay to show updated status
	go func() {
		time.Sleep(2 * time.Second)
		m.syncWorkflowHistory(m.syncWorkflowHistoryContext)
	}()
}

func (m *ModelGithubWorkflowHistory) cancelWorkflow() {
	m.status.SetProgressMessage("Canceling workflow...")

	if err := m.github.CancelWorkflow(context.Background(), gu.CancelWorkflowInput{
		Repository: m.selectedRepository.RepositoryName,
		WorkflowID: m.selectedWorkflowID,
	}); err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Failed to cancel workflow")
		return
	}

	m.status.SetSuccessMessage("Canceled workflow")
}
