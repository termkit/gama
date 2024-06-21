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
	pkgversion "github.com/termkit/gama/pkg/version"
)

type model struct {
	// current handler's properties
	TabsWithColor []string
	currentTab    *int
	isTabActive   bool

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

const (
	minTerminalWidth  = 102
	minTerminalHeight = 24
)

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	var currentTab = new(int)
	var forceUpdateWorkflowHistory = new(bool)
	var lockTabs = new(bool)

	*lockTabs = true // by default lock tabs

	tabsWithColor := []string{"Info", "Repository", "Workflow History", "Workflow", "Trigger"}

	selectedRepository := hdltypes.SelectedRepository{}

	// setup models
	hdlModelInfo := hdlinfo.SetupModelInfo(githubUseCase, version, lockTabs)
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
	m.viewport = viewport.Model{Width: minTerminalWidth, Height: minTerminalHeight}
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
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.modelInfo, cmd = m.modelInfo.Update(msg)
		cmds = append(cmds, cmd)
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

func (m *model) View() string {
	if m.viewport.Width < minTerminalWidth || m.viewport.Height < minTerminalHeight {
		return fmt.Sprintf("Terminal window is too small. Please resize to at least %dx%d.", minTerminalWidth, minTerminalHeight)
	}

	var mainDoc strings.Builder
	var helpDoc string
	var operationDoc string
	var helpDocHeight int

	header := newHeader(&m.viewport)
	for i, t := range m.TabsWithColor {
		var tabStyle lipgloss.Style
		isActive := i == *m.currentTab
		if isActive {
			tabStyle = ts.TitleStyleActive
		} else {
			if *m.lockTabs {
				tabStyle = ts.TitleStyleDisabled
			} else {
				tabStyle = ts.TitleStyleInactive
			}
		}
		header.addHeaderTab(t, tabStyle)
	}

	mainDoc.WriteString("\n" + header.renderHeader() + "\n")

	var width = lipgloss.Width(strings.Repeat("-", m.viewport.Width)) - 4
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

// --- Header ---

// header is a helper for rendering the header of the terminal.
type header struct {
	viewport *viewport.Model
	titles   []string
}

// newHeader returns a new header.
func newHeader(viewport *viewport.Model) *header {
	return &header{
		titles:   make([]string, 0),
		viewport: viewport,
	}
}

// addHeaderTab adds a tab to the header.
func (h *header) addHeaderTab(title string, style lipgloss.Style) {
	h.titles = append(h.titles, style.Render(title))
}

// renderHeader renders the header.
func (h *header) renderHeader() string {
	line := strings.Repeat("─", max(0, h.viewport.Width-79))

	return lipgloss.JoinHorizontal(lipgloss.Center, append(h.titles, line)...)
}
