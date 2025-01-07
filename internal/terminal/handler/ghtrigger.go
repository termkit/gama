package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/gama/pkg/workflow"
	"github.com/termkit/skeleton"
	"golang.org/x/exp/slices"
)

// -----------------------------------------------------------------------------
// Model Definition
// -----------------------------------------------------------------------------

type ModelGithubTrigger struct {
	// Core dependencies
	skeleton *skeleton.Skeleton
	github   gu.UseCase

	// UI Components
	help         help.Model
	Keys         githubTriggerKeyMap
	tableTrigger table.Model
	status       *ModelStatus
	textInput    textinput.Model

	// Workflow state
	workflowContent        *workflow.Pretty
	selectedWorkflow       string
	selectedRepositoryName string
	isTriggerable          bool

	// Table state
	tableReady bool

	// Option state
	optionInit    bool
	optionCursor  int
	optionValues  []string
	currentOption string

	// Input state
	triggerFocused bool

	// Context management
	syncWorkflowContext context.Context
	cancelSyncWorkflow  context.CancelFunc

	// Shared state
	selectedRepository *SelectedRepository

	// Track last branch for refresh
	lastBranch string
}

// -----------------------------------------------------------------------------
// Constructor & Initialization
// -----------------------------------------------------------------------------

func SetupModelGithubTrigger(s *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubTrigger {
	modelStatus := SetupModelStatus(s)

	m := &ModelGithubTrigger{
		// Initialize core dependencies
		skeleton: s,
		github:   githubUseCase,

		// Initialize UI components
		help:         help.New(),
		Keys:         githubTriggerKeys,
		status:       modelStatus,
		textInput:    setupTriggerInput(),
		tableTrigger: setupTriggerTable(),

		// Initialize state
		selectedRepository:  NewSelectedRepository(),
		syncWorkflowContext: context.Background(),
		cancelSyncWorkflow:  func() {},
	}

	return m
}

func setupTriggerInput() textinput.Model {
	ti := textinput.New()
	ti.Blur()
	ti.CharLimit = 72
	return ti
}

func setupTriggerTable() table.Model {
	t := table.New(
		table.WithColumns(tableColumnsTrigger),
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

	return t
}

// -----------------------------------------------------------------------------
// Bubbletea Model Implementation
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (m *ModelGithubTrigger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmd := m.handleWorkflowChange(); cmd != nil {
		return m, cmd
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle key messages
	switch msg := msg.(type) {
	case tea.KeyMsg:
		cmd = m.handleKeyMsg(msg)
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

func (m *ModelGithubTrigger) View() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		m.renderTable(),
		m.renderInputArea(),
		m.status.View(),
		m.renderHelp(),
	)
}

// -----------------------------------------------------------------------------
// Workflow Change Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) handleWorkflowChange() tea.Cmd {
	if m.selectedRepository.WorkflowName == "" {
		m.handleNoWorkflow()
		return nil
	}

	if m.shouldSyncWorkflow() {
		m.lastBranch = m.selectedRepository.BranchName
		return m.initializeWorkflowSync()
	}

	return nil
}

func (m *ModelGithubTrigger) handleNoWorkflow() {
	m.status.Reset()
	m.status.SetDefaultMessage("No workflow selected.")
	m.fillTableWithEmptyMessage()
}

func (m *ModelGithubTrigger) shouldSyncWorkflow() bool {
	return m.selectedRepository.WorkflowName != "" &&
		(m.selectedRepository.WorkflowName != m.selectedWorkflow ||
			m.selectedRepository.RepositoryName != m.selectedRepositoryName ||
			m.lastBranch != m.selectedRepository.BranchName)
}

func (m *ModelGithubTrigger) initializeWorkflowSync() tea.Cmd {
	m.tableReady = false
	m.isTriggerable = false
	m.triggerFocused = false

	m.cancelSyncWorkflow()

	m.selectedWorkflow = m.selectedRepository.WorkflowName
	m.selectedRepositoryName = m.selectedRepository.RepositoryName
	m.syncWorkflowContext, m.cancelSyncWorkflow = context.WithCancel(context.Background())

	go m.syncWorkflowContent(m.syncWorkflowContext)
	return nil
}

// -----------------------------------------------------------------------------
// Key & Input Handling
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	switch {
	case msg.String() == "up":
		return m.handleUpKey()
	case msg.String() == "down":
		return m.handleDownKey()
	case key.Matches(msg, m.Keys.Refresh):
		go m.syncWorkflowContent(m.syncWorkflowContext)
	case msg.String() == "left":
		m.handleLeftKey()
	case msg.String() == "right":
		m.handleRightKey()
	case msg.String() == "tab":
		m.handleTabKey()
	case msg.String() == "enter":
		if m.triggerFocused && m.isTriggerable {
			go m.triggerWorkflow()
		}
	}
	return nil
}

func (m *ModelGithubTrigger) handleUpKey() tea.Cmd {
	if len(m.tableTrigger.Rows()) > 0 && !m.triggerFocused {
		m.tableTrigger.MoveUp(1)
		m.switchBetweenInputAndTable()
		m.optionInit = false
	}
	return nil
}

func (m *ModelGithubTrigger) handleDownKey() tea.Cmd {
	if len(m.tableTrigger.Rows()) > 0 && !m.triggerFocused {
		m.tableTrigger.MoveDown(1)
		m.switchBetweenInputAndTable()
		m.optionInit = false
	}
	return nil
}

func (m *ModelGithubTrigger) handleLeftKey() {
	if !m.triggerFocused {
		m.optionCursor = max(m.optionCursor-1, 0)
	}
}

func (m *ModelGithubTrigger) handleRightKey() {
	if !m.triggerFocused {
		m.optionCursor = min(m.optionCursor+1, len(m.optionValues)-1)
	}
}

func (m *ModelGithubTrigger) handleTabKey() {
	if m.isTriggerable {
		m.triggerFocused = !m.triggerFocused
		if m.triggerFocused {
			m.tableTrigger.Blur()
			m.textInput.Blur()
			m.showInformationIfAnyEmptyValue()
		} else {
			m.tableTrigger.Focus()
			m.textInput.Focus()
		}
	}
}

// -----------------------------------------------------------------------------
// Input Management
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) switchBetweenInputAndTable() {
	selectedRow := m.tableTrigger.SelectedRow()
	if len(selectedRow) == 0 {
		return
	}

	if selectedRow[1] == "input" || selectedRow[1] == "bool" {
		m.textInput.Focus()
		m.tableTrigger.Blur()
	} else {
		m.textInput.Blur()
		m.tableTrigger.Focus()
	}

	m.textInput.SetValue(selectedRow[4])
	m.textInput.SetCursor(len(m.textInput.Value()))
}

func (m *ModelGithubTrigger) inputController() {
	if m.workflowContent == nil {
		return
	}

	selectedRow := m.tableTrigger.SelectedRow()
	if len(selectedRow) == 0 {
		return
	}

	switch selectedRow[1] {
	case "choice", "bool":
		m.handleChoiceInput(selectedRow)
	default:
		m.handleTextInput()
	}
}

func (m *ModelGithubTrigger) handleChoiceInput(row []string) {
	var optionValues []string
	if row[1] == "choice" {
		optionValues = m.getChoiceValues(row[0])
	} else {
		optionValues = m.getBooleanValues(row[0])
	}

	m.optionValues = optionValues
	if !m.optionInit {
		for i, option := range m.optionValues {
			if option == row[4] {
				m.optionCursor = i
			}
		}
	}
	m.optionInit = true

	state := &InputState{
		Type:    row[1],
		Options: m.optionValues,
		Cursor:  m.optionCursor,
	}
	m.updateInputState(state)
}

func (m *ModelGithubTrigger) handleTextInput() {
	m.optionValues = nil
	m.optionCursor = 0

	if !m.triggerFocused {
		m.textInput.Focus()
	}

	if m.textInput.Focused() {
		state := &InputState{
			Type:      "input",
			Value:     m.textInput.Value(),
			IsFocused: true,
		}
		m.updateInputState(state)
	}
}

// -----------------------------------------------------------------------------
// Workflow Content Management
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) syncWorkflowContent(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s@%s] Fetching workflow contents...",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))

	m.tableTrigger.SetRows([]table.Row{})

	workflowContent, err := m.github.InspectWorkflow(ctx, gu.InspectWorkflowInput{
		Repository:   m.selectedRepository.RepositoryName,
		Branch:       m.selectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
	})

	if err != nil {
		m.handleWorkflowError(err)
		return
	}

	m.processWorkflowContent(workflowContent)
}

func (m *ModelGithubTrigger) processWorkflowContent(content *gu.InspectWorkflowOutput) {
	if content.Workflow == nil {
		m.status.SetError(errors.New("workflow contents cannot be empty"))
		m.status.SetErrorMessage("You have no workflow contents")
		return
	}

	m.workflowContent = content.Workflow
	m.updateTriggerTable()
	m.finalizeWorkflowUpdate()
}

// -----------------------------------------------------------------------------
// Table Management
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) updateTriggerTable() {
	var rows []table.Row

	// Add key-value inputs
	for _, keyVal := range m.workflowContent.KeyVals {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", keyVal.ID),
			"input",
			keyVal.Key,
			keyVal.Default,
			keyVal.Value,
		})
	}

	// Add choice inputs
	for _, choice := range m.workflowContent.Choices {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", choice.ID),
			"choice",
			choice.Key,
			choice.Default,
			choice.Value,
		})
	}

	// Add regular inputs
	for _, input := range m.workflowContent.Inputs {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", input.ID),
			"input",
			input.Key,
			input.Default,
			input.Value,
		})
	}

	// Add boolean inputs
	for _, boolean := range m.workflowContent.Boolean {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", boolean.ID),
			"bool",
			boolean.Key,
			boolean.Default,
			boolean.Value,
		})
	}

	m.tableTrigger.SetRows(rows)
	m.sortTableItemsByName()
}

func (m *ModelGithubTrigger) sortTableItemsByName() {
	rows := m.tableTrigger.Rows()
	slices.SortFunc(rows, func(a, b table.Row) int {
		return strings.Compare(a[2], b[2])
	})
	m.tableTrigger.SetRows(rows)
}

func (m *ModelGithubTrigger) finalizeWorkflowUpdate() {
	m.tableTrigger.SetCursor(0)
	m.optionCursor = 0
	m.optionValues = nil
	m.triggerFocused = false
	m.tableTrigger.Focus()

	m.textInput.SetCursor(0)
	m.textInput.SetValue("")
	m.textInput.Placeholder = ""

	m.tableReady = true
	m.isTriggerable = true

	if m.hasNoWorkflowOptions() {
		m.handleEmptyWorkflow()
	} else {
		m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow contents fetched.",
			m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
	}
}

// -----------------------------------------------------------------------------
// Trigger Logic
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) triggerWorkflow() {
	if m.triggerFocused {
		m.fillEmptyValuesWithDefault()
	}

	m.status.SetProgressMessage(fmt.Sprintf("[%s@%s]:[%s] Triggering workflow...",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName, m.selectedWorkflow))

	if m.workflowContent == nil {
		m.status.SetErrorMessage("Workflow contents cannot be empty")
		return
	}

	content, err := m.workflowContent.ToJson()
	if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Workflow contents cannot be converted to JSON")
		return
	}

	if err := m.github.TriggerWorkflow(context.Background(), gu.TriggerWorkflowInput{
		Repository:   m.selectedRepository.RepositoryName,
		Branch:       m.selectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
		Content:      content,
	}); err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Workflow cannot be triggered")
		return
	}

	m.handleSuccessfulTrigger()
}

func (m *ModelGithubTrigger) handleSuccessfulTrigger() {
	m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s]:[%s] Workflow triggered.",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName, m.selectedWorkflow))

	m.status.SetProgressMessage("Switching to workflow history tab...")
	time.Sleep(2000 * time.Millisecond)

	m.resetTriggerState()
	m.switchToHistoryTab()
}

// -----------------------------------------------------------------------------
// UI Rendering
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) renderTable() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		MarginLeft(1)

	if m.triggerFocused {
		baseStyle = baseStyle.BorderForeground(lipgloss.Color("240"))
	} else {
		baseStyle = baseStyle.BorderForeground(lipgloss.Color("#3b698f"))
	}

	m.updateTableDimensions()
	return baseStyle.Render(m.tableTrigger.View())
}

func (m *ModelGithubTrigger) renderInputArea() string {
	var selectedRow = m.tableTrigger.SelectedRow()
	var selector = m.emptySelector()

	if len(m.tableTrigger.Rows()) > 0 {
		if selectedRow[1] == "input" {
			selector = m.inputSelector()
		} else {
			selector = m.optionSelector()
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, selector, m.triggerButton())
}

func (m *ModelGithubTrigger) renderHelp() string {
	helpStyle := WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)
	return helpStyle.Render(m.help.View(m.Keys))
}

// -----------------------------------------------------------------------------
// Helper Functions - Selectors
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) emptySelector() string {
	return m.getSelectorStyle().Render("")
}

func (m *ModelGithubTrigger) inputSelector() string {
	return m.getSelectorStyle().Render(m.textInput.View())
}

func (m *ModelGithubTrigger) optionSelector() string {
	var processedValues []string
	for i, option := range m.optionValues {
		if i == m.optionCursor {
			processedValues = append(processedValues, m.getSelectedOptionStyle().Render(option))
		} else {
			processedValues = append(processedValues, m.getUnselectedOptionStyle().Render(option))
		}
	}

	return m.getSelectorStyle().Render(lipgloss.JoinHorizontal(lipgloss.Left, processedValues...))
}

func (m *ModelGithubTrigger) triggerButton() string {
	button := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("255")).
		Padding(0, 1).
		Align(lipgloss.Center)

	if m.triggerFocused {
		button = button.
			BorderForeground(lipgloss.Color("#399adb")).
			Foreground(lipgloss.Color("#399adb")).
			BorderStyle(lipgloss.DoubleBorder())
	}

	return button.Render("Trigger")
}

// -----------------------------------------------------------------------------
// Helper Functions - Styles
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) getSelectorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 17).
		MarginLeft(1)
}

func (m *ModelGithubTrigger) getSelectedOptionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("120")).
		Padding(0, 1)
}

func (m *ModelGithubTrigger) getUnselectedOptionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("140")).
		Padding(0, 1)
}

// -----------------------------------------------------------------------------
// Helper Functions - Value Management
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) getChoiceValues(id string) []string {
	for _, choice := range m.workflowContent.Choices {
		if fmt.Sprintf("%d", choice.ID) == id {
			return choice.Values
		}
	}
	return nil
}

func (m *ModelGithubTrigger) getBooleanValues(id string) []string {
	for _, boolean := range m.workflowContent.Boolean {
		if fmt.Sprintf("%d", boolean.ID) == id {
			return boolean.Values
		}
	}
	return nil
}

func (m *ModelGithubTrigger) updateChoiceValue(row []string, value string) {
	rows := m.tableTrigger.Rows()
	for i, r := range rows {
		if r[0] == row[0] {
			rows[i][4] = value
		}
	}
	m.tableTrigger.SetRows(rows)

	if row[1] == "choice" {
		m.updateWorkflowChoiceValue(row[0])
	} else {
		m.updateWorkflowBooleanValue(row[0])
	}
}

func (m *ModelGithubTrigger) updateWorkflowChoiceValue(id string) {
	for i, choice := range m.workflowContent.Choices {
		if fmt.Sprintf("%d", choice.ID) == id {
			m.workflowContent.Choices[i].SetValue(m.optionValues[m.optionCursor])
			break
		}
	}
}

func (m *ModelGithubTrigger) updateWorkflowBooleanValue(id string) {
	for i, boolean := range m.workflowContent.Boolean {
		if fmt.Sprintf("%d", boolean.ID) == id {
			m.workflowContent.Boolean[i].SetValue(m.optionValues[m.optionCursor])
			break
		}
	}
}

func (m *ModelGithubTrigger) updateTextInputValue(row []string, value string) {
	if strings.HasPrefix(value, " ") {
		return
	}

	rows := m.tableTrigger.Rows()
	for i, r := range rows {
		if r[0] == row[0] {
			rows[i][4] = value
		}
	}
	m.tableTrigger.SetRows(rows)

	m.updateWorkflowInputValue(row)
}

func (m *ModelGithubTrigger) updateWorkflowInputValue(row []string) {
	for i, input := range m.workflowContent.Inputs {
		if fmt.Sprintf("%d", input.ID) == row[0] {
			m.textInput.Placeholder = input.Default
			m.workflowContent.Inputs[i].SetValue(m.textInput.Value())
			return
		}
	}

	for i, keyVal := range m.workflowContent.KeyVals {
		if fmt.Sprintf("%d", keyVal.ID) == row[0] {
			m.textInput.Placeholder = keyVal.Default
			m.workflowContent.KeyVals[i].SetValue(m.textInput.Value())
			return
		}
	}
}

// -----------------------------------------------------------------------------
// Helper Functions - State Management
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) fillEmptyValuesWithDefault() {
	if m.workflowContent == nil {
		return
	}

	rows := m.tableTrigger.Rows()
	for i, row := range rows {
		if row[4] == "" {
			rows[i][4] = rows[i][3]
		}
	}
	m.tableTrigger.SetRows(rows)

	m.fillWorkflowEmptyValues()
}

func (m *ModelGithubTrigger) fillWorkflowEmptyValues() {
	for i, choice := range m.workflowContent.Choices {
		if choice.Value == "" {
			m.workflowContent.Choices[i].SetValue(choice.Default)
		}
	}

	for i, input := range m.workflowContent.Inputs {
		if input.Value == "" {
			m.workflowContent.Inputs[i].SetValue(input.Default)
		}
	}

	for i, keyVal := range m.workflowContent.KeyVals {
		if keyVal.Value == "" {
			m.workflowContent.KeyVals[i].SetValue(keyVal.Default)
		}
	}

	for i, boolean := range m.workflowContent.Boolean {
		if boolean.Value == "" {
			m.workflowContent.Boolean[i].SetValue(boolean.Default)
		}
	}
}

func (m *ModelGithubTrigger) resetTriggerState() {
	m.workflowContent = nil
	m.selectedWorkflow = ""
	m.currentOption = ""
	m.optionValues = nil
	m.selectedRepositoryName = ""
}

func (m *ModelGithubTrigger) switchToHistoryTab() {
	m.skeleton.TriggerUpdateWithMsg(workflowHistoryUpdateMsg{time.Second * 3})
	m.skeleton.SetActivePage("history")
}

func (m *ModelGithubTrigger) hasNoWorkflowOptions() bool {
	return len(m.workflowContent.KeyVals) == 0 &&
		len(m.workflowContent.Choices) == 0 &&
		len(m.workflowContent.Inputs) == 0
}

func (m *ModelGithubTrigger) handleEmptyWorkflow() {
	m.fillTableWithEmptyMessage()
	m.status.SetDefaultMessage(fmt.Sprintf("[%s@%s] Workflow doesn't contain options but still triggerable",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
}

func (m *ModelGithubTrigger) fillTableWithEmptyMessage() {
	var rows []table.Row
	for i := 0; i < 100; i++ {
		idx := fmt.Sprintf("%d", i)
		rows = append(rows, table.Row{
			idx, "EMPTY", "EMPTY", "EMPTY", "No workflow input found",
		})
	}

	m.tableTrigger.SetRows(rows)
	m.tableTrigger.SetCursor(0)
}

func (m *ModelGithubTrigger) showInformationIfAnyEmptyValue() {
	for _, row := range m.tableTrigger.Rows() {
		if row[4] == "" {
			m.status.SetDefaultMessage("Info: You have empty values. These values uses their default values.")
			return
		}
	}
}

func (m *ModelGithubTrigger) handleWorkflowError(err error) {
	if errors.Is(err, context.Canceled) {
		return
	}
	m.status.SetError(err)
	m.status.SetErrorMessage("Workflow contents cannot be fetched")
}

// -----------------------------------------------------------------------------
// Helper Functions - Table Dimensions
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) updateTableDimensions() {
	var tableWidth int
	for _, t := range tableColumnsTrigger {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsTrigger
	widthDiff := m.skeleton.GetTerminalWidth() - tableWidth
	if widthDiff > 0 {
		keyWidth := &newTableColumns[2].Width
		valueWidth := &newTableColumns[4].Width

		*valueWidth += widthDiff - 16
		if *valueWidth%2 == 0 {
			*keyWidth = *valueWidth / 2
		}
		m.tableTrigger.SetColumns(newTableColumns)
		m.tableTrigger.SetHeight(m.skeleton.GetTerminalHeight() - 17)
	}
}

// -----------------------------------------------------------------------------
// UI Component Updates
// -----------------------------------------------------------------------------

func (m *ModelGithubTrigger) updateUIComponents(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Update text input
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	// Update table
	m.tableTrigger, cmd = m.tableTrigger.Update(msg)
	cmds = append(cmds, cmd)

	// Handle input controller
	m.inputController()

	return tea.Batch(cmds...)
}

// -----------------------------------------------------------------------------
// Input & Table State Management
// -----------------------------------------------------------------------------

type InputState struct {
	Type      string // "input", "choice", "bool"
	Value     string
	Default   string
	Options   []string
	Cursor    int
	IsFocused bool
}

func (m *ModelGithubTrigger) updateInputState(state *InputState) {
	row := m.tableTrigger.SelectedRow()
	if len(row) == 0 {
		return
	}

	switch state.Type {
	case "choice", "bool":
		m.updateChoiceValue(row, state.Options[state.Cursor])
	case "input":
		m.updateTextInputValue(row, state.Value)
	}
}
