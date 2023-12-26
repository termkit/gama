package ghrepository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	hdlerror "github.com/termkit/gama/internal/terminal/handler/error"
	"github.com/termkit/gama/internal/terminal/handler/taboptions"
	hdltypes "github.com/termkit/gama/internal/terminal/handler/types"
	"github.com/termkit/gama/pkg/browser"
)

type ModelGithubRepository struct {
	githubUseCase gu.UseCase

	Help                  help.Model
	Keys                  keyMap
	Viewport              *viewport.Model
	tableGithubRepository table.Model
	modelError            hdlerror.ModelError

	modelTabOptions       tea.Model
	actualModelTabOptions *taboptions.Options

	githubRepositories []gu.GithubRepository

	SelectedRepository *hdltypes.SelectedRepository
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func SetupModelGithubRepository(githubUseCase gu.UseCase, selectedRepository *hdltypes.SelectedRepository) *ModelGithubRepository {
	var tableRowsGithubRepository []table.Row

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
	tabOptions := taboptions.NewOptions()

	return &ModelGithubRepository{
		Help:                  help.New(),
		Keys:                  keys,
		githubUseCase:         githubUseCase,
		tableGithubRepository: tableGithubRepository,
		modelError:            modelError,
		SelectedRepository:    selectedRepository,
		modelTabOptions:       tabOptions,
		actualModelTabOptions: tabOptions,
	}
}

func (m *ModelGithubRepository) Init() tea.Cmd {
	go m.syncRepositories()

	openInBrowser := func() {
		m.modelError.SetProgressMessage(fmt.Sprintf("Opening in browser..."))

		err := browser.OpenInBrowser(fmt.Sprintf("https://github.com/%s", m.SelectedRepository.RepositoryName))
		if err != nil {
			m.modelError.SetError(err)
			m.modelError.SetErrorMessage(fmt.Sprintf("Cannot open in browser: %v", err))
			return
		}

		m.modelError.SetSuccessMessage(fmt.Sprintf("Opened in browser"))
	}

	m.actualModelTabOptions.AddOption("Open in browser", openInBrowser)

	return nil
}

func (m *ModelGithubRepository) syncRepositories() {
	m.modelError.ResetError() // reset previous errors
	m.actualModelTabOptions.SetStatus(taboptions.Wait)

	m.modelError.SetProgressMessage("Fetching repositories...")

	// delete all rows
	m.tableGithubRepository.SetRows([]table.Row{})
	m.githubRepositories = []gu.GithubRepository{}

	repositories, err := m.githubUseCase.ListRepositories(context.Background(), gu.ListRepositoriesInput{})
	if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Repositories cannot be listed")
		return
	}

	if len(repositories.Repositories) == 0 {
		m.modelError.SetDefaultMessage("No repositories found")
		return
	}

	var tableRowsGithubRepository []table.Row
	for _, repository := range repositories.Repositories {
		tableRowsGithubRepository = append(tableRowsGithubRepository,
			table.Row{repository.Name, repository.DefaultBranch, strconv.Itoa(repository.Stars), strconv.Itoa(len(repository.TriggerableWorkflows))})

		m.githubRepositories = append(m.githubRepositories, repository)
	}

	m.tableGithubRepository.SetRows(tableRowsGithubRepository)

	// set cursor to 0
	m.tableGithubRepository.SetCursor(0)

	m.modelError.SetSuccessMessage("Repositories fetched")
	m.actualModelTabOptions.SetStatus(taboptions.Idle)
	m.Update(m) // update model
}

func (m *ModelGithubRepository) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r", "R":
			go m.syncRepositories()
		}
	}
	m.tableGithubRepository, cmd = m.tableGithubRepository.Update(msg)

	// Synchronize selected repository name with parent model
	if len(m.tableGithubRepository.SelectedRow()) > 0 && m.tableGithubRepository.SelectedRow()[0] != "" {
		if m.tableGithubRepository.SelectedRow()[0] != "" {
			m.SelectedRepository.RepositoryName = m.tableGithubRepository.SelectedRow()[0]
			m.SelectedRepository.BranchName = m.tableGithubRepository.SelectedRow()[1]
		}
	}

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)

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
		newTableColumns[0].Width += widthDiff - 16
		m.tableGithubRepository.SetColumns(newTableColumns)
		m.tableGithubRepository.SetHeight(termHeight - 17)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableGithubRepository.View()))

	//optionsStyle := lipgloss.NewStyle().
	//	Foreground(lipgloss.Color("15")).
	//	Align(lipgloss.Center).Border(lipgloss.RoundedBorder())
	//
	//options := strings.Builder{}
	//
	//o1 := optionsStyle.Render("Branch > master")
	//
	//options.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, o1))

	//return lipgloss.JoinVertical(lipgloss.Top, doc.String(), options.String())

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.actualModelTabOptions.View())
}

func (m *ModelGithubRepository) ViewErrorOrOperation() string {
	return m.modelError.View()
}

func (m *ModelGithubRepository) IsError() bool {
	return m.modelError.IsError()
}
