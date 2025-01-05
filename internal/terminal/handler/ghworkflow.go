package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/termkit/gama/internal/terminal/handler/status"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/skeleton"
	"sort"

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
	mainBranch                      string

	// shared properties
	selectedRepository *hdltypes.SelectedRepository

	// use cases
	github gu.UseCase

	// keymap
	keys githubWorkflowKeyMap

	// models
	help                     help.Model
	tableTriggerableWorkflow table.Model
	status                   *status.ModelStatus
	textInput                textinput.Model
}

func SetupModelGithubWorkflow(sk *skeleton.Skeleton, githubUseCase gu.UseCase) *ModelGithubWorkflow {
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

	tableTriggerableWorkflow.KeyMap = table.KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down"),
		),
	}

	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Placeholder = "Type to switch branch"
	ti.ShowSuggestions = true

	modelStatus := status.SetupModelStatus(sk)

	return &ModelGithubWorkflow{
		skeleton:                        sk,
		help:                            help.New(),
		keys:                            githubWorkflowKeys,
		github:                          githubUseCase,
		status:                          &modelStatus,
		tableTriggerableWorkflow:        tableTriggerableWorkflow,
		selectedRepository:              hdltypes.NewSelectedRepository(),
		syncTriggerableWorkflowsContext: context.Background(),
		cancelSyncTriggerableWorkflows:  func() {},
		textInput:                       ti,
	}
}

func (m *ModelGithubWorkflow) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubWorkflow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.lastRepository != m.selectedRepository.RepositoryName {
		m.tableReady = false               // reset table ready status
		m.cancelSyncTriggerableWorkflows() // cancel previous sync
		m.syncTriggerableWorkflowsContext, m.cancelSyncTriggerableWorkflows = context.WithCancel(context.Background())

		m.lastRepository = m.selectedRepository.RepositoryName
		m.mainBranch = m.selectedRepository.BranchName

		go m.syncTriggerableWorkflows(m.syncTriggerableWorkflowsContext)
		go m.syncBranches(m.syncTriggerableWorkflowsContext)
	}

	var selectedBranch = m.textInput.Value()
	if selectedBranch != "" {
		var isBranchExist bool
		for _, branch := range m.textInput.AvailableSuggestions() {
			if branch == selectedBranch {
				isBranchExist = true
				m.selectedRepository.BranchName = selectedBranch
				break
			}
		}

		if !isBranchExist {
			m.status.SetErrorMessage(fmt.Sprintf("Branch %s is not exist", selectedBranch))
			m.skeleton.LockTabsToTheRight()
		} else {
			m.skeleton.UnlockTabs()
		}
	}
	if selectedBranch == "" {
		m.selectedRepository.BranchName = m.mainBranch
		m.skeleton.UnlockTabs()
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

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

	return lipgloss.JoinVertical(lipgloss.Top,
		style.Render(m.tableTriggerableWorkflow.View()),
		m.viewSearchBar(),
		m.status.View(),
		helpWindowStyle.Render(m.ViewHelp()),
	)
}

func (m *ModelGithubWorkflow) viewSearchBar() string {
	// Define window style
	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#3b698f")).
		Padding(0, 1).
		Width(m.skeleton.GetTerminalWidth() - 6).MarginLeft(1)

	if len(m.textInput.AvailableSuggestions()) > 0 && m.textInput.Value() == "" {
		var mainBranch = m.mainBranch
		m.textInput.Placeholder = "Type to switch branch (default: " + mainBranch + ")"
	}

	return windowStyle.Render(m.textInput.View())
}

func (m *ModelGithubWorkflow) syncTriggerableWorkflows(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s@%s] Fetching triggerable workflows...", m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))

	// delete all rows
	m.tableTriggerableWorkflow.SetRows([]table.Row{})

	triggerableWorkflows, err := m.github.GetTriggerableWorkflows(ctx, gu.GetTriggerableWorkflowsInput{
		Repository: m.selectedRepository.RepositoryName,
		Branch:     m.selectedRepository.BranchName,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Triggerable workflows cannot be listed")
		return
	}

	if len(triggerableWorkflows.TriggerableWorkflows) == 0 {
		m.selectedRepository.WorkflowName = ""
		m.status.SetDefaultMessage(fmt.Sprintf("[%s@%s] No triggerable workflow found.", m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
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

	if len(tableRowsTriggerableWorkflow) > 0 {
		m.tableTriggerableWorkflow.SetCursor(0)
	}

	m.tableReady = true
	m.status.SetSuccessMessage(fmt.Sprintf("[%s@%s] Triggerable workflows fetched.", m.selectedRepository.RepositoryName, m.selectedRepository.BranchName))
}

func (m *ModelGithubWorkflow) syncBranches(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.Reset()
	m.status.SetProgressMessage(fmt.Sprintf("[%s] Fetching branches...", m.selectedRepository.RepositoryName))

	branches, err := m.github.GetRepositoryBranches(ctx, gu.GetRepositoryBranchesInput{
		Repository: m.selectedRepository.RepositoryName,
	})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("Branches cannot be listed")
		return
	}

	if branches == nil || len(branches.Branches) == 0 {
		m.selectedRepository.BranchName = ""
		m.status.SetDefaultMessage(fmt.Sprintf("[%s] No branches found.", m.selectedRepository.RepositoryName))
		return
	}

	var bs = make([]string, len(branches.Branches))
	for i, branch := range branches.Branches {
		bs[i] = branch.Name
	}

	m.textInput.SetSuggestions(bs)

	m.status.SetSuccessMessage(fmt.Sprintf("[%s] Branches fetched.", m.selectedRepository.RepositoryName))
}

func (m *ModelGithubWorkflow) handleTableInputs(_ context.Context) {
	if !m.tableReady {
		return
	}

	// To avoid go routine leak
	rows := m.tableTriggerableWorkflow.Rows()
	selectedRow := m.tableTriggerableWorkflow.SelectedRow()

	if len(rows) > 0 && len(selectedRow) > 0 {
		m.selectedRepository.WorkflowName = selectedRow[1]
	}
}
