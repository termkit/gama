package information

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	teakey "github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
)

type ModelInfo struct {
	Help     help.Model
	Keys     keyMap
	Viewport *viewport.Model

	githubUseCase     gu.UseCase
	githubInformation githubInformation

	modelError hdlerror.ModelError
}

type githubInformation struct {
	Repositories int
	Workflows    int
	ActionRuns   int
}

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
	m.githubInformation = githubInformation{
		Repositories: 100,
		Workflows:    34,
		ActionRuns:   12,
	}
	m.modelError.SetMessage("Welcome to GAMA!")

	return nil
}

func (m *ModelInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case teakey.Matches(msg, m.Keys.Refresh):
			// Placeholder
			m.githubInformation = githubInformation{
				Repositories: rand.Intn(100),
				Workflows:    rand.Intn(100),
				ActionRuns:   rand.Intn(100),
			}
		}
	}

	return m, cmd
}

func (m *ModelInfo) View() string {
	infoDoc := strings.Builder{}

	repos := strconv.Itoa(m.githubInformation.Repositories)
	workflows := strconv.Itoa(m.githubInformation.Workflows)
	actionRuns := strconv.Itoa(m.githubInformation.ActionRuns)

	infoDoc.WriteString("Github Information\n")
	infoDoc.WriteString("Repositories: " + repos + "\n")
	infoDoc.WriteString("Workflows: " + workflows + "\n")
	infoDoc.WriteString("Action Runs: " + actionRuns + "\n")

	return infoDoc.String()
}

func (m *ModelInfo) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelInfo) IsError() bool {
	return m.modelError.IsError()
}
