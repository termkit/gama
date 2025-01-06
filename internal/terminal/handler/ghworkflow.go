package handler

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/skeleton"
)

// -----------------------------------------------------------------------------
// Model Definition
// -----------------------------------------------------------------------------

type ModelGithubWorkflow struct {
	// Core dependencies
	skeleton *skeleton.Skeleton
	github   gu.UseCase

	// UI Components
	help                     help.Model
	keys                     githubWorkflowKeyMap
	tableTriggerableWorkflow table.Model
	status                   *ModelStatus
	textInput                textinput.Model

	// Table state
	tableReady bool

	// Context management
	syncTriggerableWorkflowsContext context.Context
	cancelSyncTriggerableWorkflows  context.CancelFunc

	// Shared state
	selectedRepository *SelectedRepository

	// Indicates if there are any available workflows
	hasWorkflows           bool
	lastSelectedRepository string // Track last repository for state persistence

	// State management
	state struct {
		Ready      bool
		Repository struct {
			Current  string
			Last     string
			Branch   string
			HasFlows bool
		}
		Syncing bool
	}
}

// -----------------------------------------------------------------------------
// Constructor & Initialization
// -----------------------------------------------------------------------------

func SetupModelGithubWorkflow(s *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubWorkflow {
	m := &ModelGithubWorkflow{
		// Initialize core dependencies
		skeleton: s,
		github:   githubUseCase,

		// Initialize UI components
		help:      help.New(),
		keys:      githubWorkflowKeys,
		status:    SetupModelStatus(s),
		textInput: setupBranchInput(),

		// Initialize state
		selectedRepository:              NewSelectedRepository(),
		syncTriggerableWorkflowsContext: context.Background(),
		cancelSyncTriggerableWorkflows:  func() {},
	}

	// Setup table and blur initially
	m.tableTriggerableWorkflow = setupWorkflowTable()
	m.tableTriggerableWorkflow.Blur()
	m.textInput.Blur()

	return m
}

func setupBranchInput() textinput.Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Placeholder = "Type to switch branch"
	ti.ShowSuggestions = true
	return ti
}

func setupWorkflowTable() table.Model {
	t := table.New(
		table.WithColumns(tableColumnsWorkflow),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Apply styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Set keymap
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

func (m *ModelGithubWorkflow) Init() tea.Cmd {
	// Check initial state
	if m.lastSelectedRepository == m.selectedRepository.RepositoryName && !m.hasWorkflows {
		m.skeleton.LockTab("trigger")
		// Blur components initially
		m.tableTriggerableWorkflow.Blur()
		m.textInput.Blur()
	}
	return nil
}

func (m *ModelGithubWorkflow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check repository change and return command if exists
	if cmd := m.handleRepositoryChange(); cmd != nil {
		return m, cmd
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update text input and handle branch selection
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	m.handleBranchSelection()

	// Update table and handle workflow selection
	m.tableTriggerableWorkflow, cmd = m.tableTriggerableWorkflow.Update(msg)
	cmds = append(cmds, cmd)
	m.handleTableInputs()

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubWorkflow) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		m.renderTable(),
		m.renderBranchInput(),
		m.status.View(),
		m.renderHelp(),
	)
}

// -----------------------------------------------------------------------------
// Repository Change Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) handleRepositoryChange() tea.Cmd {
	if m.state.Repository.Current != m.selectedRepository.RepositoryName {
		m.state.Ready = false
		m.state.Repository.Current = m.selectedRepository.RepositoryName
		m.state.Repository.Branch = m.selectedRepository.BranchName
		m.syncWorkflows()
	} else if !m.state.Repository.HasFlows {
		m.skeleton.LockTab("trigger")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Branch Selection & Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) handleBranchSelection() {
	selectedBranch := m.textInput.Value()

	// Set branch
	if selectedBranch == "" {
		m.selectedRepository.BranchName = m.state.Repository.Branch
	} else if m.isBranchValid(selectedBranch) {
		m.selectedRepository.BranchName = selectedBranch
	} else {
		m.status.SetErrorMessage(fmt.Sprintf("Branch %s does not exist", selectedBranch))
		m.skeleton.LockTabsToTheRight()
		return
	}

	// Update tab state
	m.updateTabState()
}

func (m *ModelGithubWorkflow) isBranchValid(branch string) bool {
	for _, suggestion := range m.textInput.AvailableSuggestions() {
		if suggestion == branch {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Workflow Sync & Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) syncWorkflows() {
	if m.state.Syncing {
		m.cancelSyncTriggerableWorkflows()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelSyncTriggerableWorkflows = cancel
	m.state.Syncing = true

	go func() {
		defer func() {
			m.state.Syncing = false
			m.skeleton.TriggerUpdate()
		}()

		m.syncBranches(ctx)
		m.syncTriggerableWorkflows(ctx)
	}()
}

func (m *ModelGithubWorkflow) syncTriggerableWorkflows(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.initializeSyncState()
	workflows, err := m.fetchTriggerableWorkflows(ctx)
	if err != nil {
		return
	}

	m.processWorkflows(workflows)
}

func (m *ModelGithubWorkflow) initializeSyncState() {
	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s@%s] Fetching triggerable workflows...",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
	m.tableTriggerableWorkflow.SetRows([]table.Row{})
}

func (m *ModelGithubWorkflow) fetchTriggerableWorkflows(ctx context.Context) (*gu.GetTriggerableWorkflowsOutput, error) {
	workflows, err := m.github.GetTriggerableWorkflows(ctx, gu.GetTriggerableWorkflowsInput{
		Repository: m.selectedRepository.RepositoryName,
		Branch:     m.selectedRepository.BranchName,
	})

	if err != nil {
		if !errors.Is(err, context.Canceled) {
			m.status.SetError(err)
			m.status.SetErrorMessage("Triggerable workflows cannot be listed")
		}
		return nil, err
	}

	return workflows, nil
}

func (m *ModelGithubWorkflow) processWorkflows(workflows *gu.GetTriggerableWorkflowsOutput) {
	m.state.Repository.HasFlows = len(workflows.TriggerableWorkflows) > 0
	m.state.Repository.Current = m.selectedRepository.RepositoryName
	m.state.Ready = true

	if !m.state.Repository.HasFlows {
		m.handleEmptyWorkflows()
		return
	}

	m.updateWorkflowTable(workflows.TriggerableWorkflows)
	m.updateTabState()
	m.finalizeUpdate()

	// Focus components when workflows exist
	m.tableTriggerableWorkflow.Focus()
	m.textInput.Focus()
}

func (m *ModelGithubWorkflow) handleEmptyWorkflows() {
	m.selectedRepository.WorkflowName = ""
	m.skeleton.LockTab("trigger")

	// Blur components when no workflows
	m.tableTriggerableWorkflow.Blur()
	m.textInput.Blur()

	m.status.SetDefaultMessage(fmt.Sprintf("[%s@%s] No triggerable workflow found.",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))

	m.fillTableWithEmptyMessage()
}

func (m *ModelGithubWorkflow) fillTableWithEmptyMessage() {
	var rows []table.Row
	for i := 0; i < 100; i++ {
		rows = append(rows, table.Row{
			"EMPTY",
			"No triggerable workflow found",
		})
	}

	m.tableTriggerableWorkflow.SetRows(rows)
	m.tableTriggerableWorkflow.SetCursor(0)
}

// -----------------------------------------------------------------------------
// Branch Sync & Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) syncBranches(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s] Fetching branches...",
		m.selectedRepository.RepositoryName))

	branches, err := m.fetchBranches(ctx)
	if err != nil {
		return
	}

	m.processBranches(branches)
}

func (m *ModelGithubWorkflow) fetchBranches(ctx context.Context) (*gu.GetRepositoryBranchesOutput, error) {
	branches, err := m.github.GetRepositoryBranches(ctx, gu.GetRepositoryBranchesInput{
		Repository: m.selectedRepository.RepositoryName,
	})

	if err != nil {
		if !errors.Is(err, context.Canceled) {
			m.status.SetError(err)
			m.status.SetErrorMessage("Branches cannot be listed")
		}
		return nil, err
	}

	return branches, nil
}

func (m *ModelGithubWorkflow) processBranches(branches *gu.GetRepositoryBranchesOutput) {
	if branches == nil || len(branches.Branches) == 0 {
		m.handleEmptyBranches()
		return
	}

	branchNames := make([]string, len(branches.Branches))
	for i, branch := range branches.Branches {
		branchNames[i] = branch.Name
	}

	m.textInput.SetSuggestions(branchNames)
	m.status.SetSuccessMessage(fmt.Sprintf("[%s] Branches fetched.",
		m.selectedRepository.RepositoryName))
}

func (m *ModelGithubWorkflow) handleEmptyBranches() {
	m.selectedRepository.BranchName = ""
	m.status.SetDefaultMessage(fmt.Sprintf("[%s] No branches found.",
		m.selectedRepository.RepositoryName))
}

// -----------------------------------------------------------------------------
// Table Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) updateWorkflowTable(workflows []gu.TriggerableWorkflow) {
	rows := make([]table.Row, 0, len(workflows))
	for _, workflow := range workflows {
		rows = append(rows, table.Row{
			workflow.Name,
			workflow.Path,
		})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	m.tableTriggerableWorkflow.SetRows(rows)
	if len(rows) > 0 {
		m.tableTriggerableWorkflow.SetCursor(0)
	}
}

func (m *ModelGithubWorkflow) finalizeUpdate() {
	m.tableReady = true
	m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s] Triggerable workflows fetched.",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
}

func (m *ModelGithubWorkflow) handleTableInputs() {
	if !m.tableReady {
		return
	}

	rows := m.tableTriggerableWorkflow.Rows()
	selectedRow := m.tableTriggerableWorkflow.SelectedRow()
	if len(rows) > 0 && len(selectedRow) > 0 {
		m.selectedRepository.WorkflowName = selectedRow[1]
	}
}

// -----------------------------------------------------------------------------
// UI Rendering
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) renderTable() string {
	style := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		MarginLeft(1)

	m.updateTableDimensions()
	return style.Render(m.tableTriggerableWorkflow.View())
}

func (m *ModelGithubWorkflow) renderBranchInput() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 6).
		MarginLeft(1)

	if len(m.textInput.AvailableSuggestions()) > 0 && m.textInput.Value() == "" {
		if !m.state.Repository.HasFlows {
			m.textInput.Placeholder = "Branch selection disabled - No triggerable workflows available"
		} else {
			m.textInput.Placeholder = fmt.Sprintf("Type to switch branch (default: %s)", m.state.Repository.Branch)
		}
	}

	return style.Render(m.textInput.View())
}

func (m *ModelGithubWorkflow) renderHelp() string {
	helpStyle := WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)
	return helpStyle.Render(m.help.View(m.keys))
}

func (m *ModelGithubWorkflow) updateTableDimensions() {
	termWidth := m.skeleton.GetTerminalWidth()
	termHeight := m.skeleton.GetTerminalHeight()

	var tableWidth int
	for _, t := range tableColumnsWorkflow {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsWorkflow
	widthDiff := termWidth - tableWidth
	if widthDiff > 0 {
		newTableColumns[1].Width += widthDiff - 10
		m.tableTriggerableWorkflow.SetColumns(newTableColumns)
		m.tableTriggerableWorkflow.SetHeight(termHeight - 17)
	}
}

// -----------------------------------------------------------------------------
// Tab Management
// -----------------------------------------------------------------------------

func (m *ModelGithubWorkflow) updateTabState() {
	if !m.state.Repository.HasFlows {
		m.skeleton.LockTab("trigger")
		return
	}
	m.skeleton.UnlockTabs()
}
