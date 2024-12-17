package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/gama/internal/terminal/handler/status"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/gama/pkg/workflow"
	"github.com/termkit/skeleton"
	"slices"
	"strings"
	"time"
)

type ModelGithubTrigger struct {
	skeleton *skeleton.Skeleton

	// current handler's properties
	syncWorkflowContext    context.Context
	cancelSyncWorkflow     context.CancelFunc
	workflowContent        *workflow.Pretty
	tableReady             bool
	isTriggerable          bool
	optionInit             bool
	optionCursor           int
	optionValues           []string
	currentOption          string
	selectedWorkflow       string
	selectedRepositoryName string
	triggerFocused         bool

	// shared properties
	selectedRepository *hdltypes.SelectedRepository

	// use cases
	github gu.UseCase

	// keymap
	Keys githubTriggerKeyMap

	// models
	help         help.Model
	status       *status.ModelStatus
	textInput    textinput.Model
	tableTrigger table.Model
}

func SetupModelGithubTrigger(sk *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubTrigger {
	var tableRowsTrigger []table.Row

	tableTrigger := table.New(
		table.WithColumns(tableColumnsTrigger),
		table.WithRows(tableRowsTrigger),
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
	tableTrigger.SetStyles(s)

	ti := textinput.New()
	ti.Blur()
	ti.CharLimit = 72

	modelStatus := status.SetupModelStatus(sk)
	return &ModelGithubTrigger{
		skeleton:            sk,
		help:                help.New(),
		Keys:                githubTriggerKeys,
		github:              githubUseCase,
		selectedRepository:  hdltypes.NewSelectedRepository(),
		status:              &modelStatus,
		tableTrigger:        tableTrigger,
		textInput:           ti,
		syncWorkflowContext: context.Background(),
		cancelSyncWorkflow:  func() {},
	}
}

func (m *ModelGithubTrigger) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

func (m *ModelGithubTrigger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.selectedRepository.WorkflowName == "" {
		m.status.Reset()
		m.status.SetDefaultMessage("No workflow selected.")
		m.fillTableWithEmptyMessage()
		return m, nil
	}

	if m.selectedRepository.WorkflowName != "" && (m.selectedRepository.WorkflowName != m.selectedWorkflow || m.selectedRepository.RepositoryName != m.selectedRepositoryName) {
		m.tableReady = false
		m.isTriggerable = false
		m.triggerFocused = false

		m.cancelSyncWorkflow() // cancel previous sync workflow

		m.selectedWorkflow = m.selectedRepository.WorkflowName
		m.selectedRepositoryName = m.selectedRepository.RepositoryName
		m.syncWorkflowContext, m.cancelSyncWorkflow = context.WithCancel(context.Background())

		go m.syncWorkflowContent(m.syncWorkflowContext)
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch shadowMsg := msg.(type) {
	case tea.KeyMsg:
		switch shadowMsg.String() {
		case "up":
			if len(m.tableTrigger.Rows()) > 0 && !m.triggerFocused {
				m.tableTrigger.MoveUp(1)
				m.switchBetweenInputAndTable()
				// delete msg key to prevent moving cursor
				msg = tea.KeyMsg{Type: tea.KeyNull}

				m.optionInit = false
			}
		case "down":
			if len(m.tableTrigger.Rows()) > 0 && !m.triggerFocused {
				m.tableTrigger.MoveDown(1)
				m.switchBetweenInputAndTable()
				// delete msg key to prevent moving cursor
				msg = tea.KeyMsg{Type: tea.KeyNull}

				m.optionInit = false
			}
		case "ctrl+r", "ctrl+R":
			go m.syncWorkflowContent(m.syncWorkflowContext)
		case "left":
			if !m.triggerFocused {
				m.optionCursor = max(m.optionCursor-1, 0)
			}
		case "right":
			if !m.triggerFocused {
				m.optionCursor = min(m.optionCursor+1, len(m.optionValues)-1)
			}
		case "tab":
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
		case "enter", tea.KeyEnter.String():
			if m.triggerFocused && m.isTriggerable {
				go m.triggerWorkflow()
			}
		}
	}

	m.tableTrigger, cmd = m.tableTrigger.Update(msg)
	cmds = append(cmds, cmd)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.inputController(m.syncWorkflowContext)

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubTrigger) View() string {
	baseStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).MarginLeft(1)
	helpWindowStyle := hdltypes.WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)

	if m.triggerFocused {
		baseStyle = baseStyle.BorderForeground(lipgloss.Color("240"))
	} else {
		baseStyle = baseStyle.BorderForeground(lipgloss.Color("#3b698f"))
	}

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

	var selectedRow = m.tableTrigger.SelectedRow()
	var selector = m.emptySelector()
	if len(m.tableTrigger.Rows()) > 0 {
		if selectedRow[1] == "input" {
			selector = m.inputSelector()
		} else {
			selector = m.optionSelector()
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		baseStyle.Render(m.tableTrigger.View()), lipgloss.JoinHorizontal(lipgloss.Top, selector, m.triggerButton()),
		m.status.View(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelGithubTrigger) switchBetweenInputAndTable() {
	var selectedRow = m.tableTrigger.SelectedRow()

	if selectedRow[1] == "input" || selectedRow[1] == "bool" {
		m.textInput.Focus()
		m.tableTrigger.Blur()
	} else {
		m.textInput.Blur()
		m.tableTrigger.Focus()
	}
	m.textInput.SetValue(m.tableTrigger.SelectedRow()[4])
	m.textInput.SetCursor(len(m.textInput.Value()))
}

func (m *ModelGithubTrigger) inputController(_ context.Context) {
	if m.workflowContent == nil {
		return
	}

	if len(m.tableTrigger.Rows()) > 0 {
		var selectedRow = m.tableTrigger.SelectedRow()
		if len(selectedRow) == 0 {
			return
		}

		switch selectedRow[1] {
		case "choice":
			var optionValues []string
			for _, choice := range m.workflowContent.Choices {
				if fmt.Sprintf("%d", choice.ID) == selectedRow[0] {
					optionValues = append(optionValues, choice.Values...)
				}
			}
			m.optionValues = optionValues
			if !m.optionInit {
				for i, option := range m.optionValues {
					if option == selectedRow[4] {
						m.optionCursor = i
					}
				}
			}
			m.optionInit = true
		case "bool":
			var optionValues []string
			for _, choice := range m.workflowContent.Boolean {
				if fmt.Sprintf("%d", choice.ID) == selectedRow[0] {
					optionValues = append(optionValues, choice.Values...)
				}
			}
			m.optionValues = optionValues
			if !m.optionInit {
				for i, option := range m.optionValues {
					if option == selectedRow[4] {
						m.optionCursor = i
					}
				}
			}
			m.optionInit = true
		default:
			m.optionValues = nil
			m.optionCursor = 0

			if !m.triggerFocused {
				m.textInput.Focus()
			}
		}
	}

	for i, choice := range m.workflowContent.Choices {
		var selectedRow = m.tableTrigger.SelectedRow()
		var rows = m.tableTrigger.Rows()

		if len(selectedRow) == 0 || len(rows) == 0 {
			return
		}
		if fmt.Sprintf("%d", choice.ID) == selectedRow[0] {
			m.workflowContent.Choices[i].SetValue(m.optionValues[m.optionCursor])

			for i, row := range rows {
				if row[0] == selectedRow[0] {
					rows[i][4] = m.optionValues[m.optionCursor]
				}
			}

			m.tableTrigger.SetRows(rows)
		}
	}

	if m.workflowContent.Boolean != nil {
		for i, boolean := range m.workflowContent.Boolean {
			var selectedRow = m.tableTrigger.SelectedRow()
			var rows = m.tableTrigger.Rows()
			if len(selectedRow) == 0 || len(rows) == 0 {
				return
			}
			if fmt.Sprintf("%d", boolean.ID) == selectedRow[0] {
				m.workflowContent.Boolean[i].SetValue(m.optionValues[m.optionCursor])

				for i, row := range rows {
					if row[0] == selectedRow[0] {
						rows[i][4] = m.optionValues[m.optionCursor]
					}
				}

				m.tableTrigger.SetRows(rows)
			}
		}
	}

	if m.textInput.Focused() {
		if strings.HasPrefix(m.textInput.Value(), " ") {
			m.textInput.SetValue("")
		}

		var selectedRow = m.tableTrigger.SelectedRow()
		var rows = m.tableTrigger.Rows()
		if len(selectedRow) == 0 || len(rows) == 0 {
			return
		}

		for i, input := range m.workflowContent.Inputs {
			if fmt.Sprintf("%d", input.ID) == selectedRow[0] {
				m.textInput.Placeholder = input.Default
				m.workflowContent.Inputs[i].SetValue(m.textInput.Value())

				for i, row := range rows {
					if row[0] == selectedRow[0] {
						rows[i][4] = m.textInput.Value()
					}
				}

				m.tableTrigger.SetRows(rows)
			}
		}

		for i, keyVal := range m.workflowContent.KeyVals {
			if fmt.Sprintf("%d", keyVal.ID) == selectedRow[0] {
				m.textInput.Placeholder = keyVal.Default
				m.workflowContent.KeyVals[i].SetValue(m.textInput.Value())

				for i, row := range rows {
					if row[0] == selectedRow[0] {
						rows[i][4] = m.textInput.Value()
					}
				}

				m.tableTrigger.SetRows(rows)
			}
		}
	}
}

func (m *ModelGithubTrigger) syncWorkflowContent(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.Reset()
	m.status.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching workflow contents...",
			m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))

	// reset table rows
	m.tableTrigger.SetRows([]table.Row{})

	workflowContent, err := m.github.InspectWorkflow(ctx, gu.InspectWorkflowInput{
		Repository:   m.selectedRepository.RepositoryName,
		Branch:       m.selectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Workflow contents cannot be fetched")
		return
	}

	if workflowContent.Workflow == nil {
		m.status.SetError(errors.New("workflow contents cannot be empty"))
		m.status.SetErrorMessage("You have no workflow contents")
		return
	}

	m.workflowContent = workflowContent.Workflow

	var tableRowsTrigger []table.Row
	for _, keyVal := range m.workflowContent.KeyVals {
		tableRowsTrigger = append(tableRowsTrigger, table.Row{
			fmt.Sprintf("%d", keyVal.ID),
			"input", // json type
			keyVal.Key,
			keyVal.Default,
			keyVal.Value,
		})
	}

	for _, choice := range m.workflowContent.Choices {
		tableRowsTrigger = append(tableRowsTrigger, table.Row{
			fmt.Sprintf("%d", choice.ID),
			"choice",
			choice.Key,
			choice.Default,
			choice.Value,
		})
	}

	for _, input := range m.workflowContent.Inputs {
		tableRowsTrigger = append(tableRowsTrigger, table.Row{
			fmt.Sprintf("%d", input.ID),
			"input",
			input.Key,
			input.Default,
			input.Value,
		})
	}

	for _, boolean := range m.workflowContent.Boolean {
		tableRowsTrigger = append(tableRowsTrigger, table.Row{
			fmt.Sprintf("%d", boolean.ID),
			"bool",
			boolean.Key,
			boolean.Default,
			boolean.Value,
		})
	}

	m.tableTrigger.SetRows(tableRowsTrigger)
	m.sortTableItemsByName()
	m.tableTrigger.SetCursor(0)
	m.optionCursor = 0
	m.optionValues = nil
	m.triggerFocused = false
	m.tableTrigger.Focus()

	// reset input value
	m.textInput.SetCursor(0)
	m.textInput.SetValue("")
	m.textInput.Placeholder = ""

	m.tableReady = true
	m.isTriggerable = true

	if len(workflowContent.Workflow.KeyVals) == 0 &&
		len(workflowContent.Workflow.Choices) == 0 &&
		len(workflowContent.Workflow.Inputs) == 0 {
		m.fillTableWithEmptyMessage()
		m.status.SetDefaultMessage(fmt.Sprintf("[%s@%s] Workflow doesn't contain options but still triggerable",
			m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
	} else {
		m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow contents fetched.",
			m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
	}
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

func (m *ModelGithubTrigger) triggerButton() string {
	button := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("255")).
		Padding(0, 1).
		Align(lipgloss.Center)

	if m.triggerFocused {
		button = button.BorderForeground(lipgloss.Color("#399adb")).
			Foreground(lipgloss.Color("#399adb")).
			BorderStyle(lipgloss.DoubleBorder())
	}

	return button.Render("Trigger")
}

func (m *ModelGithubTrigger) fillEmptyValuesWithDefault() {
	if m.workflowContent == nil {
		m.status.SetError(errors.New("workflow contents cannot be empty"))
		m.status.SetErrorMessage("You have no workflow contents")
		return
	}

	rows := m.tableTrigger.Rows()
	for i, row := range rows {
		if row[4] == "" {
			rows[i][4] = rows[i][3]
		}
	}
	m.tableTrigger.SetRows(rows)

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

	_, err = m.github.TriggerWorkflow(context.Background(), gu.TriggerWorkflowInput{
		Repository:   m.selectedRepository.RepositoryName,
		Branch:       m.selectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
		Content:      content,
	})
	if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Workflow cannot be triggered")
		return
	}

	m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s]:[%s] Workflow triggered.",
		m.selectedRepository.RepositoryName, m.selectedRepository.BranchName, m.selectedWorkflow))

	m.skeleton.TriggerUpdate()
	m.status.SetProgressMessage("Switching to workflow history tab...")
	time.Sleep(2000 * time.Millisecond)

	// move these operations under new function named "resetTabSettings"
	m.workflowContent = nil       // reset workflow content
	m.selectedWorkflow = ""       // reset selected workflow
	m.currentOption = ""          // reset current option
	m.optionValues = nil          // reset option values
	m.selectedRepositoryName = "" // reset selected repository name

	m.skeleton.TriggerUpdateWithMsg(workflowHistoryUpdateMsg{time.Second * 3}) // update workflow history
	m.skeleton.SetActivePage("history")                                        // switch tab to workflow history
}

func (m *ModelGithubTrigger) emptySelector() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 18).MarginLeft(1)

	return windowStyle.Render("")
}

func (m *ModelGithubTrigger) inputSelector() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 18).MarginLeft(1)

	return windowStyle.Render(m.textInput.View())
}

// optionSelector renders the options list
// TODO: Make this dynamic limited&sized.
func (m *ModelGithubTrigger) optionSelector() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 18).MarginLeft(1)

	// Define styles for selected and unselected options
	selectedOptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Padding(0, 1)
	unselectedOptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("140")).Padding(0, 1)

	var processedValues []string
	for i, option := range m.optionValues {
		if i == m.optionCursor {
			processedValues = append(processedValues, selectedOptionStyle.Render(option))
		} else {
			processedValues = append(processedValues, unselectedOptionStyle.Render(option))
		}
	}

	return windowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, processedValues...))
}

func (m *ModelGithubTrigger) sortTableItemsByName() {
	rows := m.tableTrigger.Rows()
	slices.SortFunc(rows, func(a, b table.Row) int {
		return strings.Compare(a[2], b[2])
	})
	m.tableTrigger.SetRows(rows)
}
