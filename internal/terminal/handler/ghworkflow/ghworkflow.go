package ghworkflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	"github.com/termkit/gama/internal/terminal/handler/ghtrigger"
	"github.com/termkit/gama/internal/terminal/handler/taboptions"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
)

type ModelGithubWorkflow struct {
	githubUseCase gu.UseCase

	Help                     help.Model
	Keys                     keyMap
	Viewport                 *viewport.Model
	list                     list.Model
	tableTriggerableWorkflow table.Model
	modelError               hdlerror.ModelError

	modelTabOptions       tea.Model
	actualModelTabOptions *taboptions.Options

	modelGithubTrigger       tea.Model
	actualModelGithubTrigger *ghtrigger.ModelGithubTrigger

	lastRepository     string
	SelectedRepository *hdltypes.SelectedRepository
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func SetupModelGithubWorkflow(githubUseCase gu.UseCase, selectedRepository *hdltypes.SelectedRepository) *ModelGithubWorkflow {
	var tableRowsTriggerableWorkflow []table.Row

	tableTriggerableWorkflow := table.New(
		table.WithColumns(tableColumnsWorkflow),
		table.WithRows(tableRowsTriggerableWorkflow),
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
	tableTriggerableWorkflow.SetStyles(s)

	tabOptions := taboptions.NewOptions()

	return &ModelGithubWorkflow{
		Help:                     help.New(),
		Keys:                     keys,
		githubUseCase:            githubUseCase,
		tableTriggerableWorkflow: tableTriggerableWorkflow,
		SelectedRepository:       selectedRepository,
		modelTabOptions:          tabOptions,
		actualModelTabOptions:    tabOptions,
	}
}

func (m *ModelGithubWorkflow) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubWorkflow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.lastRepository != m.SelectedRepository.RepositoryName {
		go m.updateTriggerableWorkflows()
		m.lastRepository = m.SelectedRepository.RepositoryName
	}

	var cmd tea.Cmd

	m.tableTriggerableWorkflow, cmd = m.tableTriggerableWorkflow.Update(msg)

	if len(m.tableTriggerableWorkflow.Rows()) > 0 {
		m.SelectedRepository.WorkflowName = m.tableTriggerableWorkflow.SelectedRow()[1]
	}

	return m, cmd
}

func (m *ModelGithubWorkflow) View() string {
	termWidth := m.Viewport.Width
	termHeight := m.Viewport.Height

	var tableWidth int
	for _, t := range tableColumnsWorkflow {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsWorkflow
	widthDiff := termWidth - tableWidth
	if widthDiff > 0 {
		newTableColumns[1].Width += widthDiff - 11
		m.tableTriggerableWorkflow.SetColumns(newTableColumns)
		m.tableTriggerableWorkflow.SetHeight(termHeight - 17)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableTriggerableWorkflow.View()))

	return doc.String()
}

func (m *ModelGithubWorkflow) updateTriggerableWorkflows() {
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching triggerable workflows...", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
	m.actualModelTabOptions.SetStatus(taboptions.OptionWait)

	// delete all rows
	m.tableTriggerableWorkflow.SetRows([]table.Row{})

	triggerableWorkflows, err := m.githubUseCase.GetTriggerableWorkflows(context.Background(), gu.GetTriggerableWorkflowsInput{
		Repository: m.SelectedRepository.RepositoryName,
		Branch:     m.SelectedRepository.BranchName,
	})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Triggerable workflows cannot be listed")
	}

	if len(triggerableWorkflows.TriggerableWorkflows) == 0 {
		m.actualModelTabOptions.SetStatus(taboptions.OptionNone)
		m.modelError.SetDefaultMessage(fmt.Sprintf("[%s@%s] No triggerable workflow found.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
		return
	}

	var tableRowsTriggerableWorkflow []table.Row
	for _, workflow := range triggerableWorkflows.TriggerableWorkflows {
		tableRowsTriggerableWorkflow = append(tableRowsTriggerableWorkflow, table.Row{
			workflow.Name,
			workflow.Path,
		})
	}

	m.tableTriggerableWorkflow.SetRows(tableRowsTriggerableWorkflow)

	if len(m.tableTriggerableWorkflow.Rows()) > 0 {
		m.SelectedRepository.WorkflowName = m.tableTriggerableWorkflow.SelectedRow()[1]
	}

	m.actualModelTabOptions.SetStatus(taboptions.OptionIdle)
	go m.Update(m)
	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s] Triggerable workflows fetched.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))
}

func (m *ModelGithubWorkflow) ViewErrorOrOperation() string {
	return m.modelError.View()
}
