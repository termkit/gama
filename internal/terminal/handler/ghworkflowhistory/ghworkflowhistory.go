package ghworkflowhistory

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
)

type ModelGithubWorkflowHistory struct {
	Help     help.Model
	Keys     keyMap
	Viewport *viewport.Model

	githubUseCase gu.UseCase

	tableWorkflowHistory table.Model

	modelError hdlerror.ModelError
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func SetupModelGithubWorkflowHistory(githubUseCase gu.UseCase) *ModelGithubWorkflowHistory {

	tableRowsWorkflowHistory := []table.Row{
		{"Auto Deploy", "canack", "2021-01-01 12:00:00", "Success", "1m 30s"},
		{"Run Tests", "canack", "2021-01-01 12:00:00", "Success", "1m 5s"},
		{"Auto Deploy", "canack", "2021-01-01 12:00:00", "Success", "5m 11s"},
		{"Run Tests", "canack", "2021-01-01 12:00:00", "Cancelled", "50s"},
	}

	tableWorkflowHistory := table.New(
		table.WithColumns(tableColumnsWorkflowHistory),
		table.WithRows(tableRowsWorkflowHistory),
		table.WithFocused(true),
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
	tableWorkflowHistory.SetStyles(s)

	return &ModelGithubWorkflowHistory{
		Help:                 help.New(),
		Keys:                 keys,
		githubUseCase:        githubUseCase,
		tableWorkflowHistory: tableWorkflowHistory,
	}
}

func (m *ModelGithubWorkflowHistory) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubWorkflowHistory) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	//switch msg := msg.(type) {
	//case tea.KeyMsg:
	//	switch msg.String() {
	//	case "q", "ctrl+c":
	//		return m, tea.Quit
	//	}
	//}
	m.tableWorkflowHistory, cmd = m.tableWorkflowHistory.Update(msg)
	return m, cmd
}

func (m *ModelGithubWorkflowHistory) View() string {
	termWidth := m.Viewport.Width
	termHeight := m.Viewport.Height

	var tableWidth int
	for _, t := range tableColumnsWorkflowHistory {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsWorkflowHistory
	widthDiff := termWidth - tableWidth
	if widthDiff > 0 {
		newTableColumns[0].Width += widthDiff - 19
		m.tableWorkflowHistory.SetColumns(newTableColumns)
		m.tableWorkflowHistory.SetHeight(termHeight - 16)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableWorkflowHistory.View()))

	return doc.String()
}

func (m *ModelGithubWorkflowHistory) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubWorkflowHistory) IsError() bool {
	return m.modelError.IsError()
}
