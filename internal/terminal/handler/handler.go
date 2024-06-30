package handler

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	hdlgithubrepo "github.com/termkit/gama/internal/terminal/handler/ghrepository"
	hdltrigger "github.com/termkit/gama/internal/terminal/handler/ghtrigger"
	hdlWorkflow "github.com/termkit/gama/internal/terminal/handler/ghworkflow"
	hdlworkflowhistory "github.com/termkit/gama/internal/terminal/handler/ghworkflowhistory"
	"github.com/termkit/gama/internal/terminal/handler/header"
	hdlinfo "github.com/termkit/gama/internal/terminal/handler/information"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	ts "github.com/termkit/gama/internal/terminal/style"
	pkgversion "github.com/termkit/gama/pkg/version"
	"strings"
	"time"
)

type model struct {
	// models
	viewport *viewport.Model

	modelError            *hdlerror.ModelError
	modelHeader           *header.Header
	modelInfo             *hdlinfo.ModelInfo
	modelGithubRepository *hdlgithubrepo.ModelGithubRepository
	modelWorkflow         *hdlWorkflow.ModelGithubWorkflow
	modelWorkflowHistory  *hdlworkflowhistory.ModelGithubWorkflowHistory
	modelTrigger          *hdltrigger.ModelGithubTrigger

	// keymap
	keys keyMap
}

func SetupTerminal(githubUseCase gu.UseCase, version pkgversion.Version) tea.Model {
	// setup models
	hdlModelError := hdlerror.SetupModelError()
	hdlModelHeader := header.NewHeader()
	hdlModelInfo := hdlinfo.SetupModelInfo(githubUseCase, version)
	hdlModelGithubRepository := hdlgithubrepo.SetupModelGithubRepository(githubUseCase)
	hdlModelWorkflowHistory := hdlworkflowhistory.SetupModelGithubWorkflowHistory(githubUseCase)
	hdlModelWorkflow := hdlWorkflow.SetupModelGithubWorkflow(githubUseCase)
	hdlModelTrigger := hdltrigger.SetupModelGithubTrigger(githubUseCase)

	hdlModelHeader.AddCommonHeader("Info", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Repository", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Workflow History", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Workflow", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.AddCommonHeader("Trigger", ts.TitleStyleInactive, ts.TitleStyleActive)
	hdlModelHeader.SetSpecialHeader("GAMA", time.Millisecond*500, ts.TitleStyleLiveModeOn, ts.TitleStyleLiveModeOff, ts.TitleStyleDisabled)

	m := model{
		viewport:              hdltypes.NewTerminalViewport(),
		modelError:            &hdlModelError,
		modelHeader:           hdlModelHeader,
		modelInfo:             hdlModelInfo,
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
		m.modelHeader.Init(),
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
		// Handle global keybindings
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}

		m.modelHeader, cmd = m.modelHeader.Update(msg)
		cmds = append(cmds, cmd)
		cmds = append(cmds, m.handleTabContent(cmd, msg))
	case hdlinfo.UpdateSpinnerMsg:
		m.modelInfo, cmd = m.modelInfo.Update(msg)
		cmds = append(cmds, cmd)
	case header.UpdateMsg:
		m.modelHeader, cmd = m.modelHeader.Update(msg)
		cmds = append(cmds, cmd)
	case hdlworkflowhistory.UpdateWorkflowHistoryMsg:
		m.modelWorkflowHistory, cmd = m.modelWorkflowHistory.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if m.viewport.Width < hdltypes.MinTerminalWidth || m.viewport.Height < hdltypes.MinTerminalHeight {
		return fmt.Sprintf("Terminal window is too small. Please resize to at least %dx%d.", hdltypes.MinTerminalWidth, hdltypes.MinTerminalHeight)
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
