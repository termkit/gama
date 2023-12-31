package ghtrigger

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
	"github.com/termkit/gama/pkg/workflow"
)

type ModelGithubTrigger struct {
	githubUseCase gu.UseCase

	Help       help.Model
	Keys       keyMap
	Viewport   *viewport.Model
	modelError hdlerror.ModelError

	tableTrigger table.Model

	optionCursor  int
	optionValues  []string
	currentOption string

	triggerFocused bool

	workflowContent *workflow.Pretty

	selectedWorkflow       string
	selectedRepositoryName string
	SelectedRepository     *hdltypes.SelectedRepository
}

func SetupModelGithubTrigger(githubUseCase gu.UseCase, selectedRepository *hdltypes.SelectedRepository) *ModelGithubTrigger {
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

	return &ModelGithubTrigger{
		Help:               help.New(),
		Keys:               keys,
		githubUseCase:      githubUseCase,
		SelectedRepository: selectedRepository,
		modelError:         hdlerror.SetupModelError(),
		tableTrigger:       tableTrigger,
	}
}

func (m *ModelGithubTrigger) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubTrigger) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.SelectedRepository.WorkflowName != m.selectedWorkflow &&
		m.SelectedRepository.RepositoryName != m.selectedRepositoryName {
		m.selectedWorkflow = m.SelectedRepository.WorkflowName
		m.selectedRepositoryName = m.SelectedRepository.RepositoryName
		go m.syncWorkflowContent()
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			m.optionCursor = max(m.optionCursor-1, 0)
		case "right":
			m.optionCursor = min(m.optionCursor+1, len(m.optionValues)-1)
		case "tab":
			m.triggerFocused = !m.triggerFocused
			if m.triggerFocused {
				m.showInformationIfAnyEmptyValue()
				m.tableTrigger.Blur()
			} else {
				m.tableTrigger.Focus()
			}
		case "enter":
			if m.triggerFocused {
				m.triggerWorkflow()
				// switch tab to workflow history
				// reset selected options
			}
		}
	}

	m.tableTrigger, cmd = m.tableTrigger.Update(msg)
	cmds = append(cmds, cmd)

	if len(m.tableTrigger.Rows()) > 0 {
		if m.tableTrigger.SelectedRow()[1] == "choice" {
			var optionValues []string
			for _, choice := range m.workflowContent.Choices {
				if fmt.Sprintf("%d", choice.ID) == m.tableTrigger.SelectedRow()[0] {
					optionValues = append(optionValues, choice.Values...)
				}
			}
			m.optionValues = optionValues
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
	}

	return m, tea.Batch(cmds...)
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

	var json string
	var err error
	if m.workflowContent != nil {
		json, err = m.workflowContent.ToJson()
		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("Workflow contents cannot converted to JSON")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(),
		lipgloss.JoinHorizontal(lipgloss.Top, m.optionSelector(), m.triggerButton()), json)
}

func (m *ModelGithubTrigger) syncWorkflowContent() {
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching workflow contents...",
			m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))

	workflowContent, err := m.githubUseCase.InspectWorkflow(context.Background(), gu.InspectWorkflowInput{
		Repository:   m.SelectedRepository.RepositoryName,
		Branch:       m.SelectedRepository.BranchName,
		WorkflowFile: m.selectedWorkflow,
	})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Workflow contents cannot be fetched")
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

	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s] Workflow contents fetched.",
		m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
}

func (m *ModelGithubTrigger) showInformationIfAnyEmptyValue() {
	if m.triggerFocused {
		for _, row := range m.tableTrigger.Rows() {
			if row[4] == "" {
				m.modelError.SetDefaultMessage("Info: You have empty values. These values uses their default values.")
				return
			}
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

func (m *ModelGithubTrigger) triggerWorkflow() {
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
}

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
