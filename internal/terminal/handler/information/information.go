package information

import (
	"context"
	"fmt"
	"github.com/termkit/gama/internal/terminal/handler/header"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	pkgversion "github.com/termkit/gama/pkg/version"
)

type ModelInfo struct {
	version pkgversion.Version

	// use cases
	github gu.UseCase

	// models
	modelHeader *header.Header

	Help       help.Model
	Viewport   *viewport.Model
	modelError hdlerror.ModelError
	spinner    spinner.Model

	// keymap
	Keys keyMap
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

func SetupModelInfo(viewport *viewport.Model, githubUseCase gu.UseCase, version pkgversion.Version) *ModelInfo {
	modelError := hdlerror.SetupModelError()
	hdlModelHeader := header.NewHeader(viewport)

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))

	return &ModelInfo{
		Viewport:    viewport,
		modelHeader: hdlModelHeader,
		github:      githubUseCase,
		version:     version,
		Help:        help.New(),
		Keys:        keys,
		modelError:  modelError,
		spinner:     s,
	}
}

func (m *ModelInfo) Init() tea.Cmd {
	currentVersion = m.version.CurrentVersion()
	applicationDescription = fmt.Sprintf("Github Actions Manager (%s)", currentVersion)

	go m.testConnection(context.Background())
	go m.checkUpdates(context.Background())
	return nil
}

func (m *ModelInfo) checkUpdates(ctx context.Context) {
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

	go m.Update(m)
}

func (m *ModelInfo) Update(msg tea.Msg) (*ModelInfo, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m *ModelInfo) View() string {
	infoDoc := strings.Builder{}

	ws := lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		Width(m.Viewport.Width - 7)

	infoDoc.WriteString(lipgloss.JoinVertical(lipgloss.Center, applicationName, applicationDescription, newVersionAvailableMsg))

	docHeight := lipgloss.Height(infoDoc.String())
	requiredNewlinesForPadding := m.Viewport.Height - docHeight - 12

	infoDoc.WriteString(strings.Repeat("\n", max(0, requiredNewlinesForPadding)))

	return ws.Render(infoDoc.String())
}

func (m *ModelInfo) testConnection(ctx context.Context) {
	ctxWithCancel, cancel := context.WithCancel(ctx)

	// TODO: make it better
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m.spinner, _ = m.spinner.Update(m.spinner.Tick())
				m.modelError.SetProgressMessage("Checking your token " + m.spinner.View())
				time.Sleep(200 * time.Millisecond)
			}
		}
	}(ctxWithCancel)
	defer cancel()

	_, err := m.github.GetAuthUser(ctx)
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("failed to test connection, please check your token&permission")
		m.modelHeader.SetLockTabs(true)
		return
	}

	m.modelError.Reset()
	m.modelError.SetSuccessMessage("Welcome to GAMA!")
	m.modelHeader.SetLockTabs(false)

	go m.Update(m)
}

func (m *ModelInfo) ViewStatus() string {
	return m.modelError.View()
}
