package handler

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	ts "github.com/termkit/gama/internal/terminal/handler/types"
	pkgversion "github.com/termkit/gama/pkg/version"
	"github.com/termkit/skeleton"
	"strings"
)

type ModelInfo struct {
	skeleton *skeleton.Skeleton
	version  pkgversion.Version

	// use cases
	github gu.UseCase

	// models
	Help       help.Model
	modelError *hdlerror.ModelError

	// keymap
	Keys githubInformationKeyMap

	updateSelfChan chan selfUpdateMsg
}

const (
	releaseURL = "https://github.com/termkit/gama/releases"

	applicationName = `
 ..|'''.|      |     '||    ||'     |     
.|'     '     |||     |||  |||     |||    
||    ....   |  ||    |'|..'||    |  ||   
'|.    ||   .''''|.   | '|' ||   .''''|.  
''|...'|  .|.  .||. .|. | .||. .|.  .||.
`
)

var (
	currentVersion         string
	newVersionAvailableMsg string
	applicationDescription string
)

func SetupModelInfo(skeleton *skeleton.Skeleton, githubUseCase gu.UseCase, version pkgversion.Version) *ModelInfo {
	modelError := hdlerror.SetupModelError(skeleton)

	return &ModelInfo{
		skeleton:   skeleton,
		github:     githubUseCase,
		version:    version,
		Help:       help.New(),
		Keys:       githubInformationKeys,
		modelError: &modelError,
	}
}

func (m *ModelInfo) selfUpdate() {
	m.updateSelfChan <- selfUpdateMsg{}
}

func (m *ModelInfo) selfListen() tea.Cmd {
	return func() tea.Msg {
		return <-m.updateSelfChan
	}
}

func (m *ModelInfo) Init() tea.Cmd {
	currentVersion = m.version.CurrentVersion()
	applicationDescription = fmt.Sprintf("Github Actions Manager (%s)", currentVersion)

	go m.checkUpdates(context.Background())
	go m.testConnection(context.Background())

	return tea.Batch(tea.EnterAltScreen, tea.SetWindowTitle("GitHub Actions Manager (GAMA)"), m.selfListen())
}

func (m *ModelInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case selfUpdateMsg:
		_ = msg
		cmds = append(cmds, m.selfListen())
	}

	return m, tea.Batch(cmds...)
}

func (m *ModelInfo) View() string {
	infoDoc := strings.Builder{}

	helpWindowStyle := ts.WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)

	requiredNewLinesForCenter := m.skeleton.GetTerminalHeight()/2 - 11
	if requiredNewLinesForCenter < 0 {
		requiredNewLinesForCenter = 0
	}
	infoDoc.WriteString(strings.Repeat("\n", requiredNewLinesForCenter))

	infoDoc.WriteString(lipgloss.JoinVertical(lipgloss.Center, applicationName, applicationDescription, newVersionAvailableMsg))

	docHeight := lipgloss.Height(infoDoc.String())
	requiredNewlinesForPadding := m.skeleton.GetTerminalHeight() - docHeight - 12

	infoDoc.WriteString(strings.Repeat("\n", max(0, requiredNewlinesForPadding)))

	return lipgloss.JoinVertical(lipgloss.Center, infoDoc.String(), m.modelError.View(), helpWindowStyle.Render(m.ViewHelp()))
}

func (m *ModelInfo) checkUpdates(ctx context.Context) {
	defer m.selfUpdate()

	isUpdateAvailable, version, err := m.version.IsUpdateAvailable(ctx)
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("failed to check updates")
		newVersionAvailableMsg = fmt.Sprintf("failed to check updates.\nPlease visit: %s", releaseURL)
		return
	}

	if isUpdateAvailable {
		newVersionAvailableMsg = fmt.Sprintf("New version available: %s\nPlease visit: %s", version, releaseURL)
	}
}

func (m *ModelInfo) testConnection(ctx context.Context) {
	defer m.selfUpdate()

	m.modelError.SetProgressMessage("Checking your token...")
	m.skeleton.LockTabs()

	_, err := m.github.GetAuthUser(ctx)
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("failed to test connection, please check your token&permission")
		m.skeleton.LockTabs()
		return
	}

	m.modelError.Reset()
	m.modelError.SetSuccessMessage("Welcome to GAMA!")
	m.skeleton.UnlockTabs()
}
