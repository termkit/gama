package information

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
)

type ModelInfo struct {
	Help     help.Model
	Keys     keyMap
	Viewport *viewport.Model

	githubUseCase gu.UseCase

	modelError hdlerror.ModelError
}

const (
	applicationName = `
 ..|'''.|      |     '||    ||'     |     
.|'     '     |||     |||  |||     |||    
||    ....   |  ||    |'|..'||    |  ||   
'|.    ||   .''''|.   | '|' ||   .''''|.  
''|...'|  .|.  .||. .|. | .||. .|.  .||.
`
	applicatonDescription = "Github Actions Manager"
)

func SetupModelInfo(githubUseCase gu.UseCase) *ModelInfo {
	modelError := hdlerror.SetupModelError()

	return &ModelInfo{
		Help:          help.New(),
		Keys:          keys,
		githubUseCase: githubUseCase,
		modelError:    modelError,
	}
}

func (m *ModelInfo) Init() tea.Cmd {
	m.modelError.SetSuccessMessage("Welcome to GAMA!")

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

	infoDoc.WriteString(lipgloss.JoinVertical(lipgloss.Center, applicationName, applicatonDescription))

	docHeight := strings.Count(infoDoc.String(), "\n")
	requiredNewlinesForPadding := m.Viewport.Height - docHeight - 13

	infoDoc.WriteString(strings.Repeat("\n", max(0, requiredNewlinesForPadding)))

	return ws.Render(infoDoc.String())
}

func (m *ModelInfo) ViewErrorOrOperation() string {
	return m.modelError.View()
}
