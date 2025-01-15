package handler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termkit/gama/internal/github/domain"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/gama/pkg/browser"
	"github.com/termkit/skeleton"
)

// -----------------------------------------------------------------------------
// Model Definition
// -----------------------------------------------------------------------------

type ModelGithubRepository struct {
	// Core dependencies
	skeleton *skeleton.Skeleton
	github   gu.UseCase

	// UI State
	tableReady bool

	// Context management
	syncRepositoriesContext context.Context
	cancelSyncRepositories  context.CancelFunc

	// Shared state
	selectedRepository *SelectedRepository

	// UI Components
	help                        help.Model
	Keys                        githubRepositoryKeyMap
	tableGithubRepository       table.Model
	searchTableGithubRepository table.Model
	status                      *ModelStatus
	textInput                   textinput.Model
	modelTabOptions             *ModelTabOptions
}

// -----------------------------------------------------------------------------
// Constructor & Initialization
// -----------------------------------------------------------------------------

func SetupModelGithubRepository(s *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubRepository {
	modelStatus := SetupModelStatus(s)
	tabOptions := NewOptions(s, modelStatus)

	m := &ModelGithubRepository{
		// Initialize core dependencies
		skeleton: s,
		github:   githubUseCase,

		// Initialize UI components
		help:            help.New(),
		Keys:            githubRepositoryKeys,
		status:          modelStatus,
		textInput:       setupTextInput(),
		modelTabOptions: tabOptions,

		// Initialize state
		selectedRepository:      NewSelectedRepository(),
		syncRepositoriesContext: context.Background(),
		cancelSyncRepositories:  func() {},
	}

	// Setup tables
	m.tableGithubRepository = setupMainTable()
	m.searchTableGithubRepository = setupSearchTable()

	return m
}

func setupMainTable() table.Model {
	t := table.New(
		table.WithColumns(tableColumnsGithubRepository),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(13),
	)

	// Apply styles
	t.SetStyles(defaultTableStyles())

	// Apply keymap
	t.KeyMap = defaultTableKeyMap()

	return t
}

func setupSearchTable() table.Model {
	return table.New(
		table.WithColumns(tableColumnsGithubRepository),
		table.WithRows([]table.Row{}),
	)
}

func setupTextInput() textinput.Model {
	ti := textinput.New()
	ti.Blur()
	ti.CharLimit = 128
	ti.Placeholder = "Type to search repository"
	ti.ShowSuggestions = false
	return ti
}

// -----------------------------------------------------------------------------
// Bubbletea Model Implementation
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) Init() tea.Cmd {
	m.setupBrowserOption()
	go m.syncRepositories(m.syncRepositoriesContext)

	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return initSyncMsg{}
	})
}

func (m *ModelGithubRepository) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	inputMsg := msg
	switch msg := msg.(type) {
	case initSyncMsg:
		m.modelTabOptions.SetStatus(StatusIdle)
		m.tableGithubRepository.SetCursor(0)
		return m, nil
	case tea.KeyMsg:
		// Handle number keys for tab options
		if m.isNumber(msg.String()) {
			inputMsg = tea.KeyMsg{}
		}

		// Handle refresh key
		if key.Matches(msg, m.Keys.Refresh) {
			m.tableReady = false
			m.cancelSyncRepositories()
			m.syncRepositoriesContext, m.cancelSyncRepositories = context.WithCancel(context.Background())
			go m.syncRepositories(m.syncRepositoriesContext)
			return m, nil
		}

		// Handle character input for search
		if m.isCharAndSymbol(msg.Runes) {
			m.resetTableCursors()
		}
	}

	// Update text input and search functionality
	if cmd := m.updateTextInput(inputMsg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Update main table and handle row selection
	if cmd := m.updateTable(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubRepository) updateTextInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	m.updateTableRowsBySearchBar()
	return cmd
}

func (m *ModelGithubRepository) updateTable(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update main table
	m.tableGithubRepository, cmd = m.tableGithubRepository.Update(msg)
	cmds = append(cmds, cmd)

	// Update search table
	m.searchTableGithubRepository, cmd = m.searchTableGithubRepository.Update(msg)
	cmds = append(cmds, cmd)

	// Handle table selection
	m.handleTableInputs(m.syncRepositoriesContext)

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m *ModelGithubRepository) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		m.renderTable(),
		m.renderSearchBar(),
		m.modelTabOptions.View(),
		m.status.View(),
		m.renderHelp(),
	)
}

// -----------------------------------------------------------------------------
// UI Rendering
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) renderTable() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		MarginLeft(1)

	// Update table dimensions
	m.updateTableDimensions()

	return baseStyle.Render(m.tableGithubRepository.View())
}

func (m *ModelGithubRepository) renderSearchBar() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 6).
		MarginLeft(1)

	if len(m.textInput.Value()) > 0 {
		style = style.BorderForeground(lipgloss.Color("39"))
	}

	return style.Render(m.textInput.View())
}

func (m *ModelGithubRepository) renderHelp() string {
	helpStyle := WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)
	return helpStyle.Render(m.ViewHelp())
}

// -----------------------------------------------------------------------------
// Data Synchronization
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) syncRepositories(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.Reset()
	m.status.SetProgressMessage("Fetching repositories...")
	m.clearTables()

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	repos, err := m.fetchRepositories(ctx)
	if err != nil {
		m.handleFetchError(err)
		return
	}

	m.updateRepositoryData(repos)
}

func (m *ModelGithubRepository) fetchRepositories(ctx context.Context) (*gu.ListRepositoriesOutput, error) {
	return m.github.ListRepositories(ctx, gu.ListRepositoriesInput{
		Limit: 100,
		Page:  5,
		Sort:  domain.SortByUpdated,
	})
}

// -----------------------------------------------------------------------------
// Table Management
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) clearTables() {
	m.tableGithubRepository.SetRows([]table.Row{})
	m.searchTableGithubRepository.SetRows([]table.Row{})
}

func (m *ModelGithubRepository) updateTableDimensions() {
	const minTableWidth = 60 // Minimum width to maintain readability
	const tablePadding = 14  // Account for borders and margins

	var tableWidth int
	for _, t := range tableColumnsGithubRepository {
		tableWidth += t.Width
	}

	termWidth := m.skeleton.GetTerminalWidth()
	if termWidth <= minTableWidth {
		return // Prevent table from becoming too narrow
	}

	newTableColumns := make([]table.Column, len(tableColumnsGithubRepository))
	copy(newTableColumns, tableColumnsGithubRepository)

	widthDiff := termWidth - tableWidth - tablePadding
	if widthDiff > 0 {
		// Add extra width to repository name column
		newTableColumns[0].Width += widthDiff
		m.tableGithubRepository.SetColumns(newTableColumns)

		// Adjust height while maintaining some padding
		maxHeight := m.skeleton.GetTerminalHeight() - 20
		if maxHeight > 0 {
			m.tableGithubRepository.SetHeight(maxHeight)
		}
	}
}

func (m *ModelGithubRepository) resetTableCursors() {
	m.tableGithubRepository.GotoTop()
	m.tableGithubRepository.SetCursor(0)
	m.searchTableGithubRepository.GotoTop()
	m.searchTableGithubRepository.SetCursor(0)
}

// -----------------------------------------------------------------------------
// Repository Data Management
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) updateRepositoryData(repos *gu.ListRepositoriesOutput) {
	if len(repos.Repositories) == 0 {
		m.modelTabOptions.SetStatus(StatusNone)
		m.status.SetDefaultMessage("No repositories found")
		m.textInput.Blur()
		return
	}

	m.skeleton.UpdateWidgetValue("repositories", fmt.Sprintf("Repository Count: %d", len(repos.Repositories)))
	m.updateTableRows(repos.Repositories)
	m.finalizeTableUpdate()
}

func (m *ModelGithubRepository) updateTableRows(repositories []gu.GithubRepository) {
	rows := make([]table.Row, 0, len(repositories))
	for _, repo := range repositories {
		rows = append(rows, table.Row{
			repo.Name,
			repo.DefaultBranch,
			strconv.Itoa(repo.Stars),
			strconv.Itoa(len(repo.Workflows)),
		})
	}

	m.tableGithubRepository.SetRows(rows)
	m.searchTableGithubRepository.SetRows(rows)
}

func (m *ModelGithubRepository) finalizeTableUpdate() {
	m.tableGithubRepository.SetCursor(0)
	m.searchTableGithubRepository.SetCursor(0)
	m.tableReady = true
	m.textInput.Focus()
	m.status.SetSuccessMessage("Repositories fetched")

	m.skeleton.TriggerUpdateWithMsg(initSyncMsg{})
}

// -----------------------------------------------------------------------------
// Search Functionality
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) updateTableRowsBySearchBar() {
	searchValue := strings.ToLower(m.textInput.Value())
	if searchValue == "" {
		// If search is empty, restore original rows
		m.tableGithubRepository.SetRows(m.searchTableGithubRepository.Rows())
		return
	}

	rows := m.searchTableGithubRepository.Rows()
	filteredRows := make([]table.Row, 0, len(rows))

	for _, row := range rows {
		if strings.Contains(strings.ToLower(row[0]), searchValue) {
			filteredRows = append(filteredRows, row)
		}
	}

	m.tableGithubRepository.SetRows(filteredRows)
	if len(filteredRows) == 0 {
		m.clearSelectedRepository()
	}
}

func (m *ModelGithubRepository) clearSelectedRepository() {
	m.selectedRepository.RepositoryName = ""
	m.selectedRepository.BranchName = ""
	m.selectedRepository.WorkflowName = ""
}

// -----------------------------------------------------------------------------
// Input Validation & Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func (m *ModelGithubRepository) isCharAndSymbol(r []rune) bool {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_./"
	for _, c := range r {
		if strings.ContainsRune(chars, c) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Browser Integration
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) setupBrowserOption() {
	openInBrowser := func() {
		m.status.SetProgressMessage("Opening in browser...")

		url := fmt.Sprintf("https://github.com/%s", m.selectedRepository.RepositoryName)
		if err := browser.OpenInBrowser(url); err != nil {
			m.status.SetError(err)
			m.status.SetErrorMessage(fmt.Sprintf("Cannot open in browser: %v", err))
			return
		}

		m.status.SetSuccessMessage("Opened in browser")
	}

	m.modelTabOptions.AddOption("Open in browser", openInBrowser)
}

// -----------------------------------------------------------------------------
// Error Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) handleFetchError(err error) {
	if errors.Is(err, context.Canceled) {
		m.status.SetDefaultMessage("Repository fetch cancelled")
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		m.status.SetErrorMessage("Repository fetch timed out")
		return
	}

	m.status.SetError(err)
	m.status.SetErrorMessage(fmt.Sprintf("Failed to list repositories: %v", err))
}

// -----------------------------------------------------------------------------
// Table Selection Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubRepository) handleTableInputs(_ context.Context) {
	if !m.tableReady {
		return
	}

	selectedRow := m.tableGithubRepository.SelectedRow()
	if len(selectedRow) > 0 && selectedRow[0] != "" {
		m.updateSelectedRepository(selectedRow)
	}
}

func (m *ModelGithubRepository) updateSelectedRepository(row []string) {
	m.selectedRepository.RepositoryName = row[0]
	m.selectedRepository.BranchName = row[1]

	if workflowCount := row[3]; workflowCount != "" {
		m.handleWorkflowTabLocking(workflowCount)
	}
}

func (m *ModelGithubRepository) handleWorkflowTabLocking(workflowCount string) {
	count, _ := strconv.Atoi(workflowCount)
	if count == 0 {
		m.skeleton.LockTab("workflow")
		m.skeleton.LockTab("trigger")
	} else {
		m.skeleton.UnlockTab("workflow")
		m.skeleton.UnlockTab("trigger")
	}
}

// initSyncMsg is a message type used to trigger a UI update after the initial sync
type initSyncMsg struct{}
