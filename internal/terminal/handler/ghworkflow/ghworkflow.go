package ghworkflow

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
)

type ModelGithubWorkflow struct {
	Help     help.Model
	Keys     keyMap
	Viewport *viewport.Model

	githubUseCase gu.UseCase

	list list.Model

	tableWorkflowHistory table.Model

	modelError hdlerror.ModelError
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func SetupModelGithubWorkflow(githubUseCase gu.UseCase) *ModelGithubWorkflow {
	tableRowsWorkflowHistory := []table.Row{
		{"Auto Deploy", ".github/workflow/auto-deploy.yaml"},
		{"Run Tests", ".github/workflow/run-tests.yaml"},
		{"Create Release", ".github/workflow/create-release.yaml"},
	}

	tableWorkflowHistory := table.New(
		table.WithColumns(tableColumnsWorkflow),
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

	return &ModelGithubWorkflow{
		Help:                 help.New(),
		Keys:                 keys,
		githubUseCase:        githubUseCase,
		tableWorkflowHistory: tableWorkflowHistory,
	}
}

func (m *ModelGithubWorkflow) Init() tea.Cmd {
	return nil
}

func (m *ModelGithubWorkflow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *ModelGithubWorkflow) View() string {
	termWidth := m.Viewport.Width
	termHeight := m.Viewport.Height

	var tableWidth int
	for _, t := range tableColumnsWorkflow {
		tableWidth += t.Width
	}

	newTableColumns := tableColumnsWorkflow
	widthDiff := termWidth - tableWidth
	if widthDiff > 0 {
		newTableColumns[1].Width += widthDiff - 11
		m.tableWorkflowHistory.SetColumns(newTableColumns)
		m.tableWorkflowHistory.SetHeight(termHeight - 17)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableWorkflowHistory.View()))

	return doc.String()
}

func (m *ModelGithubWorkflow) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubWorkflow) IsError() bool {
	return m.modelError.IsError()
}
