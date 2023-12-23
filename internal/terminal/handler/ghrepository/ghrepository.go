package ghrepository

import (
	"context"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
)

type ModelGithubRepository struct {
	Help     help.Model
	Keys     keyMap
	Viewport *viewport.Model

	githubUseCase gu.UseCase

	tableGithubRepository table.Model
	tableWorkflow         table.Model
	tableWorkflowHistory  table.Model

	modelError hdlerror.ModelError
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func SetupModelGithubRepository(githubUseCase gu.UseCase) *ModelGithubRepository {
	tableRowsGithubRepository := []table.Row{}

	tableGithubRepository := table.New(
		table.WithColumns(tableColumnsGithubRepository),
		table.WithRows(tableRowsGithubRepository),
		table.WithFocused(true),
		table.WithHeight(13),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	tableGithubRepository.SetStyles(s)

	// setup models
	modelError := hdlerror.SetupModelError()

	return &ModelGithubRepository{
		Help:                  help.New(),
		Keys:                  keys,
		githubUseCase:         githubUseCase,
		tableGithubRepository: tableGithubRepository,
		modelError:            modelError,
	}
}

func (m *ModelGithubRepository) Init() tea.Cmd {
	workflows, err := m.githubUseCase.ListRepositories(context.Background(), gu.ListRepositoriesInput{})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Repositories cannot be listed")
		return nil
	}

	var tableRowsGithubRepository []table.Row
	for _, workflow := range workflows.Repositories {
		tableRowsGithubRepository = append(tableRowsGithubRepository,
			table.Row{workflow.Name, strconv.Itoa(workflow.Stars), strconv.Itoa(len(workflow.TriggerableWorkflows))})
	}

	m.tableGithubRepository.SetRows(tableRowsGithubRepository)

	return nil
}

func (m *ModelGithubRepository) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	//switch msg := msg.(type) {
	//case tea.KeyMsg:
	//	switch msg.String() {
	//	case "q", "ctrl+c":
	//		return m, tea.Quit
	//	}
	//}
	m.tableGithubRepository, cmd = m.tableGithubRepository.Update(msg)
	return m, cmd
}

func (m *ModelGithubRepository) View() string {
	termWidth := m.Viewport.Width
	termHeight := m.Viewport.Height

	var tableWidth int
	for _, t := range tableColumnsGithubRepository {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsGithubRepository
	widthDiff := termWidth - tableWidth
	if widthDiff > 0 {
		newTableColumns[0].Width += widthDiff - 15
		m.tableGithubRepository.SetColumns(newTableColumns)
		m.tableGithubRepository.SetHeight(termHeight - 16)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableGithubRepository.View()))
	//doc.WriteString("\n\n")

	return doc.String()
}

func (m *ModelGithubRepository) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubRepository) IsError() bool {
	return m.modelError.IsError()
}
