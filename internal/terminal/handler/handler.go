package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlgithubrepo "github.com/termkit/gama/internal/terminal/handler/ghrepository"
	hdltrigger "github.com/termkit/gama/internal/terminal/handler/ghtrigger"
	hdlWorkflow "github.com/termkit/gama/internal/terminal/handler/ghworkflow"
	hdlworkflowhistory "github.com/termkit/gama/internal/terminal/handler/ghworkflowhistory"
	hdlinfo "github.com/termkit/gama/internal/terminal/handler/information"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	ts "github.com/termkit/gama/internal/terminal/style"
	vu "github.com/termkit/gama/internal/version/usecase"
)

type model struct {
	// current handler's properties
	TabsWithColor     []string
	currentTab        *int
	isTabActive       bool
	terminalSizeReady bool

	// Shared properties
	SelectedRepository *hdltypes.SelectedRepository
	lockTabs           *bool // lockTabs will be set true if test connection fails

	// models
	viewport viewport.Model
	timer    timer.Model

	modelInfo       tea.Model
	actualModelInfo *hdlinfo.ModelInfo

	modelGithubRepository       tea.Model
	actualModelGithubRepository *hdlgithubrepo.ModelGithubRepository

	modelWorkflow       tea.Model
	directModelWorkflow *hdlWorkflow.ModelGithubWorkflow

	modelWorkflowHistory       tea.Model
	directModelWorkflowHistory *hdlworkflowhistory.ModelGithubWorkflowHistory

	modelTrigger       tea.Model
	actualModelTrigger *hdltrigger.ModelGithubTrigger

	// keymap
	keys keyMap
}

func SetupTerminal(githubUseCase gu.UseCase, versionUseCase vu.UseCase) tea.Model {
	var currentTab = new(int)
	var forceUpdateWorkflowHistory = new(bool)
	var lockTabs = new(bool)

	*lockTabs = true // by default lock tabs

	tabsWithColor := []string{"Info", "Repository", "Workflow History", "Workflow", "Trigger"}

	selectedRepository := hdltypes.SelectedRepository{}

	// setup models
	hdlModelInfo := hdlinfo.SetupModelInfo(githubUseCase, versionUseCase, lockTabs)
	hdlModelGithubRepository := hdlgithubrepo.SetupModelGithubRepository(githubUseCase, &selectedRepository)
	hdlModelWorkflowHistory := hdlworkflowhistory.SetupModelGithubWorkflowHistory(githubUseCase, &selectedRepository, forceUpdateWorkflowHistory)
	hdlModelWorkflow := hdlWorkflow.SetupModelGithubWorkflow(githubUseCase, &selectedRepository)
	hdlModelTrigger := hdltrigger.SetupModelGithubTrigger(githubUseCase, &selectedRepository, currentTab, forceUpdateWorkflowHistory)

	m := model{
		lockTabs:      lockTabs,
		currentTab:    currentTab,
		TabsWithColor: tabsWithColor,
		timer:         timer.NewWithInterval(1<<63-1, time.Millisecond*200),
		modelInfo:     hdlModelInfo, actualModelInfo: hdlModelInfo,
		SelectedRepository:    &selectedRepository,
		modelGithubRepository: hdlModelGithubRepository, actualModelGithubRepository: hdlModelGithubRepository,
		modelWorkflowHistory: hdlModelWorkflowHistory, directModelWorkflowHistory: hdlModelWorkflowHistory,
		modelWorkflow: hdlModelWorkflow, directModelWorkflow: hdlModelWorkflow,
		modelTrigger: hdlModelTrigger, actualModelTrigger: hdlModelTrigger,
		keys: keys,
	}

	hdlModelInfo.Viewport = &m.viewport
	hdlModelGithubRepository.Viewport = &m.viewport
	hdlModelWorkflowHistory.Viewport = &m.viewport
	hdlModelWorkflow.Viewport = &m.viewport
	hdlModelTrigger.Viewport = &m.viewport

	return &m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tea.SetWindowTitle("GitHub Actions Manager (GAMA)"),
		m.timer.Init(),
		m.modelInfo.Init(),
		m.modelGithubRepository.Init(),
		m.modelWorkflowHistory.Init(),
		m.modelWorkflow.Init(),
		m.modelTrigger.Init())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Sync terminal size
	m.syncTerminal(msg)

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.SwitchTabLeft):
			if !*m.lockTabs {
				*m.currentTab = max(*m.currentTab-1, 0)
			}
			cmds = append(cmds, m.handleTabContent(cmd, msg))
		case key.Matches(msg, m.keys.SwitchTabRight):
			if !*m.lockTabs {
				*m.currentTab = min(*m.currentTab+1, len(m.TabsWithColor)-1)
			}
			cmds = append(cmds, m.handleTabContent(cmd, msg))
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		default:
			cmds = append(cmds, m.handleTabContent(cmd, msg))
		}
	case timer.TickMsg:
		m.timer, cmd = m.timer.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

const (
	minTerminalWidth  = 102
	minTerminalHeight = 24
)

func (m *model) View() string {
	if !m.terminalSizeReady {
		return "Setting up..."
	}

	if m.viewport.Width < minTerminalWidth || m.viewport.Height < minTerminalHeight {
		return fmt.Sprintf("Terminal window is too small. Please resize to at least %dx%d.", minTerminalWidth, minTerminalHeight)
	}

	var mainDoc strings.Builder
	var helpDoc string
	var operationDoc string
	var helpDocHeight int

	var renderedTabs []string
	for i, t := range m.TabsWithColor {
		var style lipgloss.Style
		isActive := i == *m.currentTab
		if isActive {
			style = ts.TitleStyleActive.Copy()
		} else {
			if *m.lockTabs {
				style = ts.TitleStyleDisabled.Copy()
			} else {
				style = ts.TitleStyleInactive.Copy()
			}
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	mainDoc.WriteString("\n")
	mainDoc.WriteString(m.headerView(renderedTabs...) + "\n")

	var width = lipgloss.Width(strings.Repeat("-", m.viewport.Width)) - len(renderedTabs)
	hdltypes.ScreenWidth = &width

	dynamicWindowStyle := ts.WindowStyleCyan.Width(width).Height(m.viewport.Height - 20)

	helpWindowStyle := ts.WindowStyleHelp.Width(width)
	operationWindowStyle := lipgloss.NewStyle()

	switch *m.currentTab {
	case 0:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelInfo.View()))
		operationDoc = operationWindowStyle.Render(m.actualModelInfo.ViewStatus())
		helpDoc = helpWindowStyle.Render(m.actualModelInfo.ViewHelp())
	case 1:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelGithubRepository.View()))
		operationDoc = operationWindowStyle.Render(m.actualModelGithubRepository.ViewStatus())
		helpDoc = helpWindowStyle.Render(m.actualModelGithubRepository.ViewHelp())
	case 2:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelWorkflowHistory.View()))
		operationDoc = operationWindowStyle.Render(m.directModelWorkflowHistory.ViewStatus())
		helpDoc = helpWindowStyle.Render(m.directModelWorkflowHistory.ViewHelp())
	case 3:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelWorkflow.View()))
		operationDoc = operationWindowStyle.Render(m.directModelWorkflow.ViewStatus())
		helpDoc = helpWindowStyle.Render(m.directModelWorkflow.ViewHelp())
	case 4:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelTrigger.View()))
		operationDoc = operationWindowStyle.Render(m.actualModelTrigger.ViewStatus())
		helpDoc = helpWindowStyle.Render(m.actualModelTrigger.ViewHelp())
	}

	mainDocContent := ts.DocStyle.Render(mainDoc.String())

	mainDocHeight := strings.Count(mainDocContent, "\n")
	helpDocHeight = strings.Count(helpDoc, "\n")
	errorDocHeight := strings.Count(operationDoc, "\n")
	requiredNewlinesForPadding := m.viewport.Height - mainDocHeight - helpDocHeight - errorDocHeight
	padding := strings.Repeat("\n", max(0, requiredNewlinesForPadding))

	informationPane := lipgloss.JoinVertical(lipgloss.Top, operationDoc, helpDoc)

	return mainDocContent + padding + informationPane
}

func (m *model) syncTerminal(msg tea.Msg) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())

		if !m.terminalSizeReady {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.YPosition = headerHeight + 1
			m.terminalSizeReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}
	}
}

func (m *model) handleTabContent(cmd tea.Cmd, msg tea.Msg) tea.Cmd {
	switch *m.currentTab {
	case 0:
		m.modelInfo, cmd = m.modelInfo.Update(msg)
	case 1:
		m.modelGithubRepository, cmd = m.modelGithubRepository.Update(msg)
	case 2:
		m.modelWorkflowHistory, cmd = m.modelWorkflowHistory.Update(msg)
	case 3:
		m.modelWorkflow, cmd = m.modelWorkflow.Update(msg)
	case 4:
		m.modelTrigger, cmd = m.modelTrigger.Update(msg)
	}
	return cmd
}

func (m *model) headerView(titles ...string) string {
	var renderedTitles string
	for _, t := range titles {
		renderedTitles += t
	}
	line := strings.Repeat("─", max(0, m.viewport.Width-79))
	titles = append(titles, line)
	return lipgloss.JoinHorizontal(lipgloss.Center, titles...)
}
