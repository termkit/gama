package ghrepository

import (
	"errors"
	"math/rand"
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
	tableColumnsGithubRepository := []table.Column{
		{Title: "Repository", Width: 24},
		{Title: "Stars", Width: 6},
		{Title: "Workflows", Width: 9},
	}

	tableColumnsWorkflow := []table.Column{
		{Title: "Workflow", Width: 24},
	}

	tableColumnsWorkflowHistory := []table.Column{
		{Title: "Workflow", Width: 24},
		{Title: "Triggered", Width: 12},
		{Title: "Started At", Width: 22},
		{Title: "Status", Width: 9},
		{Title: "Duration", Width: 8},
	}

	tableRowsGithubRepository := []table.Row{
		{"canack/testrepo", "1,000", "15"},
		{"canack/anotherrepo", "321", "3"},
		{"canack/funny", "2", "1"},
		{"canack/testrepo", "0", "0"},
	}

	tableRowsWorkflow := []table.Row{
		{"Auto Deploy"},
		{"Run Tests"},
	}

	tableRowsWorkflowHistory := []table.Row{
		{"Auto Deploy", "canack", "2021-01-01 12:00:00", "Success", "1m 30s"},
		{"Run Tests", "canack", "2021-01-01 12:00:00", "Success", "1m 5s"},
		{"Auto Deploy", "canack", "2021-01-01 12:00:00", "Success", "5m 11s"},
		{"Run Tests", "canack", "2021-01-01 12:00:00", "Cancelled", "50s"},
	}

	tableGithubRepository := table.New(
		table.WithColumns(tableColumnsGithubRepository),
		table.WithRows(tableRowsGithubRepository),
		table.WithFocused(true),
		table.WithHeight(13),
	)

	tableWorkflow := table.New(
		table.WithColumns(tableColumnsWorkflow),
		table.WithRows(tableRowsWorkflow),
		table.WithFocused(false),
		table.WithHeight(3),
	)

	tableWorkflowHistory := table.New(
		table.WithColumns(tableColumnsWorkflowHistory),
		table.WithRows(tableRowsWorkflowHistory),
		table.WithFocused(false),
		table.WithHeight(7),
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
	tableWorkflow.SetStyles(s)
	tableWorkflowHistory.SetStyles(s)

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

	if rand.Intn(10)%2 == 0 {
		m.modelError.SetError(errors.New("test error"))
		m.modelError.SetErrorMessage("test error message")
	} else {
		m.modelError.Reset()
		m.modelError.SetMessage("test message")
	}
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

	return doc.String()
}

func (m *ModelGithubRepository) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubRepository) IsError() bool {
	return m.modelError.IsError()
}
