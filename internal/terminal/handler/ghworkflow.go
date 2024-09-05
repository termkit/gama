package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/termkit/skeleton"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
)

type ModelGithubWorkflow struct {
	skeleton *skeleton.Skeleton
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
	Keys githubWorkflowKeyMap

	// models
	Help                     help.Model
	tableTriggerableWorkflow table.Model
	modelError               *hdlerror.ModelError

	updateSelfChan chan any
}

func SetupModelGithubWorkflow(skeleton *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubWorkflow {
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
		BorderForeground(lipgloss.Color("#3b698f")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tableTriggerableWorkflow.SetStyles(s)

	modelError := hdlerror.SetupModelError(skeleton)

	return &ModelGithubWorkflow{
		skeleton:                        skeleton,
		Help:                            help.New(),
		Keys:                            githubWorkflowKeys,
		github:                          githubUseCase,
		modelError:                      &modelError,
		tableTriggerableWorkflow:        tableTriggerableWorkflow,
		SelectedRepository:              hdltypes.NewSelectedRepository(),
		syncTriggerableWorkflowsContext: context.Background(),
		cancelSyncTriggerableWorkflows:  func() {},
		updateSelfChan:                  make(chan any),
	}
}

func (m *ModelGithubWorkflow) selfUpdate() {
	m.updateSelfChan <- selfUpdateMsg{}
}

func (m *ModelGithubWorkflow) selfListen() tea.Cmd {
	return func() tea.Msg {
		return <-m.updateSelfChan
	}
}

func (m *ModelGithubWorkflow) Init() tea.Cmd {
	return tea.Batch(m.selfListen())
}

func (m *ModelGithubWorkflow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case selfUpdateMsg:
		_ = msg
		cmds = append(cmds, m.selfListen())
	}

	if m.lastRepository != m.SelectedRepository.RepositoryName {
		m.tableReady = false               // reset table ready status
		m.cancelSyncTriggerableWorkflows() // cancel previous sync
		m.syncTriggerableWorkflowsContext, m.cancelSyncTriggerableWorkflows = context.WithCancel(context.Background())

		m.lastRepository = m.SelectedRepository.RepositoryName

		go m.syncTriggerableWorkflows(m.syncTriggerableWorkflowsContext)
	}

	m.tableTriggerableWorkflow, cmd = m.tableTriggerableWorkflow.Update(msg)
	cmds = append(cmds, cmd)

	m.handleTableInputs(m.syncTriggerableWorkflowsContext) // update table operations

	return m, tea.Batch(cmds...)
}

func (m *ModelGithubWorkflow) View() string {
	var style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).MarginLeft(1)

	helpWindowStyle := hdltypes.WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)

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

	doc := strings.Builder{}
	doc.WriteString(style.Render(m.tableTriggerableWorkflow.View()))
	doc.WriteString("\n\n\n")

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.modelError.View(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelGithubWorkflow) syncTriggerableWorkflows(ctx context.Context) {
	defer m.selfUpdate()

	m.modelError.Reset()
	m.modelError.SetProgressMessage(fmt.Sprintf("[%s@%s] Fetching triggerable workflows...", m.SelectedRepository.RepositoryName, m.SelectedRepository.BranchName))

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
		m.SelectedRepository.WorkflowName = ""
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
