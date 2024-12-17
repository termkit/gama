package handler

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	"github.com/termkit/gama/internal/terminal/handler/status"
	ts "github.com/termkit/gama/internal/terminal/handler/types"
	pkgversion "github.com/termkit/gama/pkg/version"
	"github.com/termkit/skeleton"
	"strings"
)

type ModelInfo struct {
	skeleton *skeleton.Skeleton

	logo                   string
	releaseURL             string
	newVersionAvailableMsg string
	applicationDescription string

	version pkgversion.Version

	// use cases
	github gu.UseCase

	// models
	help   help.Model
	status *status.ModelStatus

	// keymap
	keys githubInformationKeyMap
}

func SetupModelInfo(sk *skeleton.Skeleton, githubUseCase gu.UseCase, version pkgversion.Version) *ModelInfo {
	modelStatus := status.SetupModelStatus(sk)

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

	return &ModelInfo{
		skeleton:   sk,
		releaseURL: releaseURL,
		logo:       applicationName,
		github:     githubUseCase,
		version:    version,
		help:       help.New(),
		keys:       githubInformationKeys,
		status:     &modelStatus,
	}
}

func (m *ModelInfo) Init() tea.Cmd {
	m.applicationDescription = fmt.Sprintf("Github Actions Manager (%s)", m.version.CurrentVersion())

	go m.checkUpdates(context.Background())
	go m.testConnection(context.Background())

	return tea.Batch(tea.EnterAltScreen, tea.SetWindowTitle("GitHub Actions Manager (GAMA)"))
}

func (m *ModelInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *ModelInfo) View() string {
	infoDoc := strings.Builder{}

	helpWindowStyle := ts.WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)

	requiredNewLinesForCenter := m.skeleton.GetTerminalHeight()/2 - 11
	if requiredNewLinesForCenter < 0 {
		requiredNewLinesForCenter = 0
	}

	infoDoc.WriteString(strings.Repeat("\n", requiredNewLinesForCenter))
	infoDoc.WriteString(lipgloss.JoinVertical(lipgloss.Center, m.logo, m.applicationDescription, m.newVersionAvailableMsg))

	requiredNewlinesForPadding := m.skeleton.GetTerminalHeight() - lipgloss.Height(infoDoc.String()) - 12

	infoDoc.WriteString(strings.Repeat("\n", max(0, requiredNewlinesForPadding)))

	str := lipgloss.JoinVertical(lipgloss.Center, infoDoc.String(), m.status.View(), helpWindowStyle.Render(m.ViewHelp()))

	return str
}

func (m *ModelInfo) checkUpdates(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	isUpdateAvailable, version, err := m.version.IsUpdateAvailable(ctx)
	if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("failed to check updates")
		m.newVersionAvailableMsg = fmt.Sprintf("failed to check updates.\nPlease visit: %s", m.releaseURL)
		return
	}

	if isUpdateAvailable {
		m.newVersionAvailableMsg = fmt.Sprintf("New version available: %s\nPlease visit: %s", version, m.releaseURL)
	}
}

func (m *ModelInfo) testConnection(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.SetProgressMessage("Checking your token...")
	m.skeleton.LockTabs()

	_, err := m.github.GetAuthUser(ctx)
	if err != nil {
		m.status.SetError(err)
		m.status.SetErrorMessage("failed to test connection, please check your token&permission")
		m.skeleton.LockTabs()
		return
	}

	m.status.Reset()
	m.status.SetSuccessMessage("Welcome to GAMA!")
	m.skeleton.UnlockTabs()
}
