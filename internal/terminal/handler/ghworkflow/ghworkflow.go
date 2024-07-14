package ghworkflow

import (
	"context"
	"errors"
	"fmt"
	ts "github.com/termkit/gama/internal/terminal/style"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	"github.com/termkit/gama/internal/terminal/handler/ghtrigger"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
)

type ModelGithubWorkflow struct {
	// current handler's properties
	syncTriggerableWorkflowsContext context.Context
	cancelSyncTriggerableWorkflows  context.CancelFunc
	tableReady                      bool
	lastRepository                  string

	// shared properties
	SelectedRepository *hdltypes.SelectedRepository

	// use cases
	github gu.UseCase

	// keymap
	Keys keyMap

	// models
	Help                     help.Model
	Viewport                 *viewport.Model
	list                     list.Model
	tableTriggerableWorkflow table.Model
	modelError               *hdlerror.ModelError

	modelGithubTrigger *ghtrigger.ModelGithubTrigger
}

func SetupModelGithubWorkflow(githubUseCase gu.UseCase) *ModelGithubWorkflow {
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

	modelError := hdlerror.SetupModelError()

	return &ModelGithubWorkflow{
		Viewport:                        hdltypes.NewTerminalViewport(),
		Help:                            help.New(),
		Keys:                            keys,
		github:                          githubUseCase,
		modelError:                      &modelError,
		tableTriggerableWorkflow:        tableTriggerableWorkflow,
		SelectedRepository:              hdltypes.NewSelectedRepository(),
		syncTriggerableWorkflowsContext: context.Background(),
		cancelSyncTriggerableWorkflows:  func() {},
	}
}

func (m *ModelGithubWorkflow) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubWorkflow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.lastRepository != m.SelectedRepository.RepositoryName {
		m.tableReady = false               // reset table ready status
		m.cancelSyncTriggerableWorkflows() // cancel previous sync
		m.syncTriggerableWorkflowsContext, m.cancelSyncTriggerableWorkflows = context.WithCancel(context.Background())

		m.lastRepository = m.SelectedRepository.RepositoryName

		go m.syncTriggerableWorkflows(m.syncTriggerableWorkflowsContext)
	}

	m.tableTriggerableWorkflow, cmd = m.tableTriggerableWorkflow.Update(msg)

	m.handleTableInputs(m.syncTriggerableWorkflowsContext) // update table operations

	return m, cmd
}

func (m *ModelGithubWorkflow) View() string {
	var style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	helpWindowStyle := ts.WindowStyleHelp.Width(m.Viewport.Width - 4)

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
	doc.WriteString(style.Render(m.tableTriggerableWorkflow.View()))
	doc.WriteString("\n\n\n")

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.ViewStatus(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelGithubWorkflow) syncTriggerableWorkflows(ctx context.Context) {
	m.modelError.Reset()
	m.modelError.SetProgressMessage(
		fmt.Sprintf("[%s@%s] Fetching triggerable workflows...", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))

	// delete all rows
	m.tableTriggerableWorkflow.SetRows([]table.Row{})

	triggerableWorkflows, err := m.github.GetTriggerableWorkflows(ctx, gu.GetTriggerableWorkflowsInput{
		Repository: m.SelectedRepository.RepositoryName,
		Branch:     m.SelectedRepository.BranchName,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Triggerable workflows cannot be listed")
		return
	}

	if len(triggerableWorkflows.TriggerableWorkflows) == 0 {
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

	sort.SliceStable(tableRowsTriggerableWorkflow, func(i, j int) bool {
		return tableRowsTriggerableWorkflow[i][0] < tableRowsTriggerableWorkflow[j][0]
	})

	m.tableTriggerableWorkflow.SetRows(tableRowsTriggerableWorkflow)

	m.tableReady = true
	m.modelError.SetSuccessMessage(fmt.Sprintf("[%s@%s] Triggerable workflows fetched.", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))

	go m.Update(m) // update model
}

func (m *ModelGithubWorkflow) handleTableInputs(_ context.Context) {
	if !m.tableReady {
		return
	}

	// To avoid go routine leak
	rows := m.tableTriggerableWorkflow.Rows()
	selectedRow := m.tableTriggerableWorkflow.SelectedRow()

	if len(rows) > 0 && len(selectedRow) > 0 {
		m.SelectedRepository.WorkflowName = selectedRow[1]
	}
}

func (m *ModelGithubWorkflow) ViewStatus() string {
	return m.modelError.View()
}
