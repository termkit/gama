package ghrepository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	// current handler's properties
	syncRepositoriesContext context.Context
	cancelSyncRepositories  context.CancelFunc
	tableReady              bool

	// shared properties
	SelectedRepository *hdltypes.SelectedRepository

	// use cases
	githubUseCase gu.UseCase

	// keymap
	Keys keyMap

	// models
	Help                  help.Model
	Viewport              *viewport.Model
	tableGithubRepository table.Model
	modelError            *hdlerror.ModelError

	modelTabOptions       tea.Model
	actualModelTabOptions *taboptions.Options
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
	tabOptions := taboptions.NewOptions(&modelError)

	return &ModelGithubRepository{
		Help:                    help.New(),
		Keys:                    keys,
		githubUseCase:           githubUseCase,
		tableGithubRepository:   tableGithubRepository,
		modelError:              &modelError,
		SelectedRepository:      selectedRepository,
		modelTabOptions:         tabOptions,
		actualModelTabOptions:   tabOptions,
		syncRepositoriesContext: context.Background(),
		cancelSyncRepositories:  func() {},
	}
}

func (m *ModelGithubRepository) Init() tea.Cmd {
	go m.syncRepositories(m.syncRepositoriesContext)

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

func (m *ModelGithubRepository) syncRepositories(ctx context.Context) {
	m.modelError.ResetError() // reset previous errors
	m.actualModelTabOptions.SetStatus(taboptions.OptionWait)
	m.modelError.SetProgressMessage("Fetching repositories...")

	// delete all rows
	m.tableGithubRepository.SetRows([]table.Row{})

	repositories, err := m.githubUseCase.ListRepositories(ctx, gu.ListRepositoriesInput{})
	if errors.Is(err, context.Canceled) {
		return
	} else if err != nil {
		m.modelError.SetError(err)
		m.modelError.SetErrorMessage("Repositories cannot be listed")
		return
	}

	if len(repositories.Repositories) == 0 {
		m.actualModelTabOptions.SetStatus(taboptions.OptionNone)
		m.modelError.SetDefaultMessage("No repositories found")
		return
	}

	tableRowsGithubRepository := make([]table.Row, 0, len(repositories.Repositories))
	for _, repository := range repositories.Repositories {
		tableRowsGithubRepository = append(tableRowsGithubRepository,
			table.Row{repository.Name, repository.DefaultBranch, strconv.Itoa(repository.Stars), strconv.Itoa(len(repository.Workflows))})
	}

	m.tableGithubRepository.SetRows(tableRowsGithubRepository)

	// set cursor to 0
	m.tableGithubRepository.SetCursor(0)

	m.tableReady = true
	m.modelError.SetSuccessMessage("Repositories fetched")
	go m.Update(m) // update model
}

func (m *ModelGithubRepository) handleTableInputs(ctx context.Context) {
	if !m.tableReady {
		return
	}

	// To avoid go routine leak
	selectedRow := m.tableGithubRepository.SelectedRow()

	// Synchronize selected repository name with parent model
	if len(selectedRow) > 0 && selectedRow[0] != "" {
		m.SelectedRepository.RepositoryName = selectedRow[0]
		m.SelectedRepository.BranchName = selectedRow[1]
	}

	m.actualModelTabOptions.SetStatus(taboptions.OptionIdle)
}

func (m *ModelGithubRepository) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Refresh):
			m.tableReady = false       // reset table ready status
			m.cancelSyncRepositories() // cancel previous sync
			m.syncRepositoriesContext, m.cancelSyncRepositories = context.WithCancel(context.Background())
			go m.syncRepositories(m.syncRepositoriesContext)
		}
	}

	m.tableGithubRepository, cmd = m.tableGithubRepository.Update(msg)
	cmds = append(cmds, cmd)

	m.modelTabOptions, cmd = m.modelTabOptions.Update(msg)
	cmds = append(cmds, cmd)

	m.handleTableInputs(m.syncRepositoriesContext)

	return m, tea.Batch(cmds...)
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
		m.tableGithubRepository.SetHeight(termHeight - 17)
	}

	doc := strings.Builder{}
	doc.WriteString(baseStyle.Render(m.tableGithubRepository.View()))

	return lipgloss.JoinVertical(lipgloss.Top, doc.String(), m.actualModelTabOptions.View())
}

func (m *ModelGithubRepository) ViewStatus() string {
	return m.modelError.View()
}
