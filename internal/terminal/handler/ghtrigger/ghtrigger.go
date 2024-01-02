package ghtrigger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/gama/pkg/workflow"
)

type ModelGithubTrigger struct {
	githubUseCase gu.UseCase

	Help       help.Model
	Keys       keyMap
	Viewport   *viewport.Model
	modelError hdlerror.ModelError

	textInput    textinput.Model
	tableTrigger table.Model

	currentTab                 *int
	forceUpdateWorkflowHistory *bool

	optionInit    bool
	optionCursor  int
	optionValues  []string
	currentOption string

	triggerFocused bool

	workflowContent *workflow.Pretty

	selectedWorkflow       string
	selectedRepositoryName string
	SelectedRepository     *hdltypes.SelectedRepository
}

func SetupModelGithubTrigger(githubUseCase gu.UseCase, selectedRepository *hdltypes.SelectedRepository, currentTab *int, forceUpdateWorkflowHistory *bool) *ModelGithubTrigger {
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

	return &ModelGithubTrigger{
		currentTab:                 currentTab,
		forceUpdateWorkflowHistory: forceUpdateWorkflowHistory,
		Help:                       help.New(),
		Keys:                       keys,
		githubUseCase:              githubUseCase,
		SelectedRepository:         selectedRepository,
		modelError:                 hdlerror.SetupModelError(),
		tableTrigger:               tableTrigger,
		textInput:                  ti,
	}
}

func (m *ModelGithubTrigger) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ModelGithubTrigger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.SelectedRepository.WorkflowName != m.selectedWorkflow ||
		m.SelectedRepository.RepositoryName != m.selectedRepositoryName {
		m.selectedWorkflow = m.SelectedRepository.WorkflowName
		m.selectedRepositoryName = m.SelectedRepository.RepositoryName
		go m.syncWorkflowContent()
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch shadowMsg := msg.(type) {
	case tea.KeyMsg:
		switch shadowMsg.String() {
		case "up":
			if len(m.tableTrigger.Rows()) > 0 {
				m.tableTrigger.MoveUp(1)
				m.switchBetweenInputAndTable()
				// delete msg key to prevent moving cursor
				msg = tea.KeyMsg{Type: tea.KeyNull}

				m.optionInit = false
			}
		case "down":
			if len(m.tableTrigger.Rows()) > 0 {
				m.tableTrigger.MoveDown(1)
				m.switchBetweenInputAndTable()
				// delete msg key to prevent moving cursor
				msg = tea.KeyMsg{Type: tea.KeyNull}

				m.optionInit = false
			}
		case "ctrl+r", "ctrl+R":
			go m.syncWorkflowContent()
		case "left":
			m.optionCursor = max(m.optionCursor-1, 0)
		case "right":
			m.optionCursor = min(m.optionCursor+1, len(m.optionValues)-1)
		case "tab":
			m.triggerFocused = !m.triggerFocused
			if m.triggerFocused {
				m.tableTrigger.Blur()
				m.showInformationIfAnyEmptyValue()
			} else {
				m.tableTrigger.Focus()
			}
		case "enter":
			if m.triggerFocused {
				go m.triggerWorkflow()
			}
		}
	}

	m.tableTrigger, cmd = m.tableTrigger.Update(msg)
	cmds = append(cmds, cmd)

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.inputController()

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubTrigger) switchBetweenInputAndTable() {
	if m.tableTrigger.SelectedRow()[1] == "input" {
		m.textInput.Focus()
		m.tableTrigger.Blur()
	} else {
		m.textInput.Blur()
		m.tableTrigger.Focus()
	}
	m.textInput.SetValue(m.tableTrigger.SelectedRow()[4])
	m.textInput.SetCursor(len(m.textInput.Value()))
}

func (m *ModelGithubTrigger) inputController() {
	if len(m.tableTrigger.Rows()) > 0 {
		var selectedRow = m.tableTrigger.SelectedRow()
		if selectedRow[1] == "choice" {
			var optionValues []string
			for _, choice := range m.workflowContent.Choices {
				if fmt.Sprintf("%d", choice.ID) == m.tableTrigger.SelectedRow()[0] {
					optionValues = append(optionValues, choice.Values...)
				}
			}
			m.optionValues = optionValues
			if m.optionInit == false {
				for i, option := range m.optionValues {
					if option == selectedRow[4] {
						m.optionCursor = i
					}
				}
			}
			m.optionInit = true
		} else {
			m.optionValues = nil
			m.optionCursor = 0
		}
	}

	if m.workflowContent != nil {
		for i, choice := range m.workflowContent.Choices {
			if fmt.Sprintf("%d", choice.ID) == m.tableTrigger.SelectedRow()[0] {
				m.workflowContent.Choices[i].SetValue(m.optionValues[m.optionCursor])

				rows := m.tableTrigger.Rows()
				for i, row := range rows {
					if row[0] == m.tableTrigger.SelectedRow()[0] {
						rows[i][4] = m.optionValues[m.optionCursor]
					}
				}

				m.tableTrigger.SetRows(rows)
			}
		}

		if m.textInput.Focused() {
			for i, input := range m.workflowContent.Inputs {
				if fmt.Sprintf("%d", input.ID) == m.tableTrigger.SelectedRow()[0] {
					m.textInput.Placeholder = input.Default
					m.workflowContent.Inputs[i].SetValue(m.textInput.Value())

					rows := m.tableTrigger.Rows()
					for i, row := range rows {
						if row[0] == m.tableTrigger.SelectedRow()[0] {
							rows[i][4] = m.textInput.Value()
						}
					}

					m.tableTrigger.SetRows(rows)
				}
			}

			for i, keyVal := range m.workflowContent.KeyVals {
				if fmt.Sprintf("%d", keyVal.ID) == m.tableTrigger.SelectedRow()[0] {
					m.textInput.Placeholder = keyVal.Default
					m.workflowContent.KeyVals[i].SetValue(m.textInput.Value())

					rows := m.tableTrigger.Rows()
					for i, row := range rows {
						if row[0] == m.tableTrigger.SelectedRow()[0] {
							rows[i][4] = m.textInput.Value()
						}
					}

					m.tableTrigger.SetRows(rows)
				}
			}
		}
	}
}

func (m *ModelGithubTrigger) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	termWidth := m.Viewport.Width
	termHeight := m.Viewport.Height

	var tableWidth int
	for _, t := range tableColumnsTrigger {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsTrigger
	widthDiff := termWidth - tableWidth
	if widthDiff > 0 {
		newTableColumns[4].Width += widthDiff - 17
		m.tableTrigger.SetColumns(newTableColumns)
		m.tableTrigger.SetHeight(termHeight - 17)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableTrigger.View()))

	var selector string
	if len(m.tableTrigger.Rows()) > 0 {
		if m.tableTrigger.SelectedRow()[1] == "input" {
			selector = m.inputSelector()
		} else {
			selector = m.optionSelector()
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(),
		lipgloss.JoinHorizontal(lipgloss.Top, selector, m.triggerButton()))
}

func (m *ModelGithubTrigger) syncWorkflowContent() {
	m.modelError.Reset()
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching workflow contents...",
			m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))

	// reset table rows
	m.tableTrigger.SetRows([]table.Row{})

	workflowContent, err := m.githubUseCase.InspectWorkflow(context.Background(), gu.InspectWorkflowInput{
		Repository:   m.SelectedRepository.RepositoryName,
		Branch:       m.SelectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
	})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow contents cannot be fetched")
		return
	}

	if workflowContent.Workflow == nil {
		m.modelError.SetError(errors.New("workflow contents cannot be empty"))
		m.modelError.SetErrorMessage("You have no workflow contents")
		return
	}

	if len(workflowContent.Workflow.KeyVals) == 0 &&
		len(workflowContent.Workflow.Choices) == 0 &&
		len(workflowContent.Workflow.Inputs) == 0 {
		m.modelError.SetDefaultMessage(fmt.Sprintf("[%s@%s] No workflow contents found.",
			m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
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

	m.tableTrigger.SetRows(tableRowsTrigger)

	m.tableTrigger.SetCursor(0)
	m.optionCursor = 0
	m.optionValues = nil
	m.triggerFocused = false
	m.tableTrigger.Focus()

	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow contents fetched.",
		m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
}

func (m *ModelGithubTrigger) showInformationIfAnyEmptyValue() {
	for _, row := range m.tableTrigger.Rows() {
		if row[4] == "" {
			m.modelError.SetDefaultMessage("Info: You have empty values. These values uses their default values.")
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
		button = button.Copy().
			BorderForeground(lipgloss.Color("130")).
			Foreground(lipgloss.Color("130")).
			BorderStyle(lipgloss.DoubleBorder())
	}

	return button.Render("Trigger")
}

func (m *ModelGithubTrigger) fillEmptyValuesWithDefault() {
	if m.workflowContent == nil {
		m.modelError.SetError(errors.New("workflow contents cannot be empty"))
		m.modelError.SetErrorMessage("You have no workflow contents")
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
}

func (m *ModelGithubTrigger) triggerWorkflow() {
	if m.triggerFocused {
		m.fillEmptyValuesWithDefault()
	}

	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s]:[%s] Triggering workflow...",
			m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName, m.selectedWorkflow))

	if m.workflowContent == nil {
		m.modelError.SetErrorMessage("Workflow contents cannot be empty")
		return
	}

	content, err := m.workflowContent.ToJson()
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow contents cannot be converted to JSON")
		return
	}

	_, err = m.githubUseCase.TriggerWorkflow(context.Background(), gu.TriggerWorkflowInput{
		Repository:   m.SelectedRepository.RepositoryName,
		Branch:       m.SelectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
		Content:      content,
	})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow cannot be triggered")
		return
	}

	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s]:[%s] Workflow triggered.",
		m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName, m.selectedWorkflow))

	time.Sleep(1 * time.Second)
	m.modelError.SetProgressMessage("Switching to workflow history tab...")
	time.Sleep(1 * time.Second)

	// move these operations under new function named "resetTabSettings"
	m.workflowContent = nil       // reset workflow content
	m.selectedWorkflow = ""       // reset selected workflow
	m.currentOption = ""          // reset current option
	m.optionValues = nil          // reset option values
	m.selectedRepositoryName = "" // reset selected repository name

	go func() {
		time.Sleep(1 * time.Second)
		*m.forceUpdateWorkflowHistory = true // force update workflow history
	}()
	*m.currentTab = 2 // switch tab to workflow history
}

func (m *ModelGithubTrigger) inputSelector() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(*hdltypes.ScreenWidth - 13)

	// Build the options list
	doc := strings.Builder{}

	doc.WriteString(m.textInput.View())

	return windowStyle.Render(doc.String())
}

// optionSelector renders the options list
// TODO: Make this dynamic limited&sized.
func (m *ModelGithubTrigger) optionSelector() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(0, 1).
		Width(*hdltypes.ScreenWidth - 13)

	// Define styles for selected and unselected options
	selectedOptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("120")).Padding(0, 1)
	unselectedOptionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("140")).Padding(0, 1)

	// Build the options list
	doc := strings.Builder{}

	var processedValues []string
	for i, option := range m.optionValues {
		if i == m.optionCursor {
			processedValues = append(processedValues, selectedOptionStyle.Render(option))
		} else {
			processedValues = append(processedValues, unselectedOptionStyle.Render(option))
		}
	}

	horizontal := lipgloss.JoinHorizontal(lipgloss.Left, processedValues...)

	doc.WriteString(horizontal)

	// Apply window style to the entire list
	return windowStyle.Render(doc.String())
}

func (m *ModelGithubTrigger) ViewErrorOrOperation() string {
	return m.modelError.View()
}
