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
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
)

type ModelInfo struct {
	// use cases
	githubUseCase gu.UseCase

	// lockTabs will be set true if test connection fails
	lockTabs *bool

	// models
	Help       help.Model
	Viewport   *viewport.Model
	modelError hdlerror.ModelError
	spinner    spinner.Model

	// keymap
	Keys keyMap
}

const (
	applicationName = `
 ..|'''.|      |     '||    ||'     |     
.|'     '     |||     |||  |||     |||    
||    ....   |  ||    |'|..'||    |  ||   
'|.    ||   .''''|.   | '|' ||   .''''|.  
''|...'|  .|.  .||. .|. | .||. .|.  .||.
`
)

var (
	GamaVersion            = "under development" // will be set by build flag
	applicationDescription = fmt.Sprintf("Github Actions Manager (%s)", GamaVersion)
)

func SetupModelInfo(githubUseCase gu.UseCase, lockTabs *bool) *ModelInfo {
	modelError := hdlerror.SetupModelError()

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))

	return &ModelInfo{
		Help:          help.New(),
		Keys:          keys,
		githubUseCase: githubUseCase,
		modelError:    modelError,
		lockTabs:      lockTabs,
		spinner:       s,
	}
}

func (m *ModelInfo) Init() tea.Cmd {
	go m.testConnection(context.Background())

	return nil
}

func (m *ModelInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	ws := lipgloss.NewStyle().BorderForeground(lipgloss.Color("39")).Align(lipgloss.Center).Border(lipgloss.RoundedBorder()).Width(m.Viewport.Width - 7)

	infoDoc.WriteString(lipgloss.JoinVertical(lipgloss.Center, applicationName, applicationDescription))

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

	_, err := m.githubUseCase.ListRepositories(ctx, gu.ListRepositoriesInput{Limit: 1})
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
