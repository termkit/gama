package information

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	pkgversion "github.com/termkit/gama/pkg/version"
)

type ModelInfo struct {
	version pkgversion.Version

	// use cases
	github gu.UseCase

	// lockTabs will be set true if test connection fails
	lockTabs *bool

	// models
	Help              help.Model
	Viewport          *viewport.Model
	modelError        hdlerror.ModelError
	spinner           spinner.Model
	changelogViewer   *viewport.Model
	changelogRenderer *glamour.TermRenderer

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

func SetupModelInfo(githubUseCase gu.UseCase, version pkgversion.Version, lockTabs *bool) *ModelInfo {
	modelError := hdlerror.SetupModelError()

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))

	vp := viewport.New(96, 6)
	vp.Style = lipgloss.NewStyle()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(96),
	)
	if err != nil {
		panic(err)
	}

	return &ModelInfo{
		github:            githubUseCase,
		version:           version,
		Help:              help.New(),
		Keys:              keys,
		modelError:        modelError,
		changelogRenderer: renderer,
		changelogViewer:   &vp,
		lockTabs:          lockTabs,
		spinner:           s,
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
		newVersionAvailableMsg = fmt.Sprintf("failed to check updates: %v\nPlease visit: %s", err, releaseURL)
		return
	}

	if isUpdateAvailable {
		var changelogView strings.Builder

		changelogView.WriteString("```markdown\n")

		changelogView.WriteString(fmt.Sprintf("GAMA has new version! %s\n\n", version))

		changeLogs, err := m.version.ChangelogsSinceCurrentVersion(ctx)
		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("failed to get changelogs")
			return
		}

		changelogView.WriteString(fmt.Sprintf("\n# Changelogs"))

		for _, cl := range changeLogs {
			publishedAt := cl.PublishedAt.Format("2006-01-02 15:04:05")
			changelogView.WriteString(fmt.Sprintf("\n\n## %s - %s\n\n%s", cl.TagName, publishedAt, cl.Body))
			changelogView.WriteString("\n\n")
			changelogView.WriteString("##### ")
			changelogView.WriteString(strings.Repeat("-", 84))
			changelogView.WriteString("\n")
		}

		changelogView.WriteString("\n```")
		changelogView.WriteString("\n")

		str, err := m.changelogRenderer.Render(changelogView.String())
		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage("failed to render changelogs")
			return
		}

		//newVersionAvailableMsg = str

		m.changelogViewer.SetContent(str)
	}

	go m.Update(m)
}

func (m *ModelInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Help.Width = msg.Width
		m.changelogViewer.Width = msg.Width - 2
		m.changelogViewer.Height = msg.Height - 20
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			*m.changelogViewer, cmd = m.changelogViewer.Update(msg)
			return m, cmd
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

	infoDoc.WriteString(lipgloss.JoinVertical(lipgloss.Center, applicationName, applicationDescription))
	infoDoc.WriteString("\n")
	infoDoc.WriteString(m.changelogViewer.View())

	docHeight := strings.Count(infoDoc.String(), "\n")
	requiredNewlinesForPadding := m.Viewport.Height - docHeight - 13

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
		*m.lockTabs = true
		return
	}

	m.modelError.Reset()
	m.modelError.SetSuccessMessage("Welcome to GAMA!")
	*m.lockTabs = false

	go m.Update(m)
}

func (m *ModelInfo) ViewStatus() string {
	return m.modelError.View()
}
