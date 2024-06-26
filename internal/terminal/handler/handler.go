package handler

import (
	"fmt"
	"github.com/termkit/gama/internal/terminal/handler/header"
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
	// Shared properties
	SelectedRepository *hdltypes.SelectedRepository

	// models
	viewport *viewport.Model
	timer    timer.Model

	modelHeader           *header.Header
	modelInfo             *hdlinfo.ModelInfo
	modelGithubRepository *hdlgithubrepo.ModelGithubRepository
	modelWorkflow         *hdlWorkflow.ModelGithubWorkflow
	modelWorkflowHistory  *hdlworkflowhistory.ModelGithubWorkflowHistory
	modelTrigger          *hdltrigger.ModelGithubTrigger

	// keymap
	keys keyMap
}

const (
	minTerminalWidth  = 102
	minTerminalHeight = 24
)

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	forceUpdateWorkflowHistory := new(bool)
	vp := &viewport.Model{Width: minTerminalWidth, Height: minTerminalHeight}
	selectedRepository := &hdltypes.SelectedRepository{}

	// setup models
	hdlModelHeader := header.NewHeader(vp)
	hdlModelInfo := hdlinfo.SetupModelInfo(vp, githubUseCase, version)
	hdlModelGithubRepository := hdlgithubrepo.SetupModelGithubRepository(vp, githubUseCase, selectedRepository)
	hdlModelWorkflowHistory := hdlworkflowhistory.SetupModelGithubWorkflowHistory(vp, githubUseCase, selectedRepository, forceUpdateWorkflowHistory)
	hdlModelWorkflow := hdlWorkflow.SetupModelGithubWorkflow(vp, githubUseCase, selectedRepository)
	hdlModelTrigger := hdltrigger.SetupModelGithubTrigger(vp, githubUseCase, selectedRepository, forceUpdateWorkflowHistory)

	hdlModelHeader.AddCommonHeader("Info", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Repository", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Workflow History", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Workflow", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Trigger", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.SetSpecialHeader("GAMA", ts.TitleStyleLiveModeOn, ts.TitleStyleLiveModeOff)

	m := model{
		viewport:              vp,
		timer:                 timer.NewWithInterval(1<<63-1, time.Millisecond*200),
		modelHeader:           hdlModelHeader,
		modelInfo:             hdlModelInfo,
		SelectedRepository:    selectedRepository,
		modelGithubRepository: hdlModelGithubRepository,
		modelWorkflowHistory:  hdlModelWorkflowHistory,
		modelWorkflow:         hdlModelWorkflow,
		modelTrigger:          hdlModelTrigger,
		keys:                  keys,
	}

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
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
	case tea.KeyMsg:
		m.modelHeader, cmd = m.modelHeader.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.handleTabContent(cmd, msg))

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
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

	var width = lipgloss.Width(strings.Repeat("-", m.viewport.Width)) - 5
	hdltypes.ScreenWidth = &width

	dynamicWindowStyle := lipgloss.NewStyle().Width(width).Height(m.viewport.Height - 22)
	helpWindowStyle := ts.WindowStyleHelp.Width(width)

	mainDoc.WriteString(m.modelHeader.View() + "\n")
	switch m.modelHeader.GetCurrentTab() {
	case 0:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelInfo.View()))
		operationDoc = m.modelInfo.ViewStatus()
		helpDoc = helpWindowStyle.Render(m.modelInfo.ViewHelp())
	case 1:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelGithubRepository.View()))
		operationDoc = m.modelGithubRepository.ViewStatus()
		helpDoc = helpWindowStyle.Render(m.modelGithubRepository.ViewHelp())
	case 2:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelWorkflowHistory.View()))
		operationDoc = m.modelWorkflowHistory.ViewStatus()
		helpDoc = helpWindowStyle.Render(m.modelWorkflowHistory.ViewHelp())
	case 3:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelWorkflow.View()))
		operationDoc = m.modelWorkflow.ViewStatus()
		helpDoc = helpWindowStyle.Render(m.modelWorkflow.ViewHelp())
	case 4:
		mainDoc.WriteString(dynamicWindowStyle.Render(m.modelTrigger.View()))
		operationDoc = m.modelTrigger.ViewStatus()
		helpDoc = helpWindowStyle.Render(m.modelTrigger.ViewHelp())
	}

	mainDocContent := ts.DocStyle.Render(mainDoc.String())
	informationPane := lipgloss.JoinVertical(lipgloss.Top, operationDoc, helpDoc)

	return lipgloss.JoinVertical(lipgloss.Top, mainDocContent, informationPane)
}

func (m *model) handleTabContent(cmd tea.Cmd, msg tea.Msg) tea.Cmd {
	switch m.modelHeader.GetCurrentTab() {
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
