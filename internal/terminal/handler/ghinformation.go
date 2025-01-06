package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gu "github.com/termkit/gama/internal/github/usecase"
	pkgversion "github.com/termkit/gama/pkg/version"
	"github.com/termkit/skeleton"
)

// -----------------------------------------------------------------------------
// Model Definition
// -----------------------------------------------------------------------------

type ModelInfo struct {
	// Core dependencies
	skeleton *skeleton.Skeleton
	github   gu.UseCase
	version  pkgversion.Version

	// UI Components
	help   help.Model
	status *ModelStatus
	keys   githubInformationKeyMap

	// Application state
	logo                   string
	releaseURL             string
	applicationDescription string
	newVersionAvailableMsg string
}

// -----------------------------------------------------------------------------
// Constructor & Initialization
// -----------------------------------------------------------------------------

func SetupModelInfo(s *skeleton.Skeleton, githubUseCase gu.UseCase, version pkgversion.Version) *ModelInfo {
	const releaseURL = "https://github.com/termkit/gama/releases"

	return &ModelInfo{
		// Initialize core dependencies
		skeleton: s,
		github:   githubUseCase,
		version:  version,

		// Initialize UI components
		help:   help.New(),
		status: SetupModelStatus(s),
		keys:   githubInformationKeys,

		// Initialize application state
		logo:       defaultLogo,
		releaseURL: releaseURL,
	}
}

const defaultLogo = `
 ..|'''.|      |     '||    ||'     |     
.|'     '     |||     |||  |||     |||    
||    ....   |  ||    |'|..'||    |  ||   
'|.    ||   .''''|.   | '|' ||   .''''|.  
''|...'|  .|.  .||. .|. | .||. .|.  .||.
`

// -----------------------------------------------------------------------------
// Bubbletea Model Implementation
// -----------------------------------------------------------------------------

func (m *ModelInfo) Init() tea.Cmd {
	m.initializeAppDescription()
	m.startBackgroundTasks()

	return tea.Batch(
		tea.EnterAltScreen,
		tea.SetWindowTitle("GitHub Actions Manager (GAMA)"),
	)
}

func (m *ModelInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *ModelInfo) View() string {
	return lipgloss.JoinVertical(lipgloss.Center,
		m.renderMainContent(),
		m.status.View(),
		m.renderHelpWindow(),
	)
}

// -----------------------------------------------------------------------------
// UI Rendering
// -----------------------------------------------------------------------------

func (m *ModelInfo) renderMainContent() string {
	content := strings.Builder{}

	// Add vertical centering
	centerPadding := m.calculateCenterPadding()
	content.WriteString(strings.Repeat("\n", centerPadding))

	// Add main content
	content.WriteString(lipgloss.JoinVertical(lipgloss.Center,
		m.logo,
		m.applicationDescription,
		m.newVersionAvailableMsg,
	))

	// Add bottom padding
	bottomPadding := m.calculateBottomPadding(content.String())
	content.WriteString(strings.Repeat("\n", bottomPadding))

	return content.String()
}

func (m *ModelInfo) renderHelpWindow() string {
	helpStyle := WindowStyleHelp.Width(m.skeleton.GetTerminalWidth() - 4)
	return helpStyle.Render(m.ViewHelp())
}

// -----------------------------------------------------------------------------
// Layout Calculations
// -----------------------------------------------------------------------------

func (m *ModelInfo) calculateCenterPadding() int {
	padding := m.skeleton.GetTerminalHeight()/2 - 11
	return max(0, padding)
}

func (m *ModelInfo) calculateBottomPadding(content string) int {
	padding := m.skeleton.GetTerminalHeight() - lipgloss.Height(content) - 12
	return max(0, padding)
}

// -----------------------------------------------------------------------------
// Application State Management
// -----------------------------------------------------------------------------

func (m *ModelInfo) initializeAppDescription() {
	m.applicationDescription = fmt.Sprintf("Github Actions Manager (%s)", m.version.CurrentVersion())
}

func (m *ModelInfo) startBackgroundTasks() {
	go m.checkUpdates(context.Background())
	go m.testConnection(context.Background())
}

// -----------------------------------------------------------------------------
// Background Tasks
// -----------------------------------------------------------------------------

func (m *ModelInfo) checkUpdates(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	isUpdateAvailable, version, err := m.version.IsUpdateAvailable(ctx)
	if err != nil {
		m.handleUpdateError(err)
		return
	}

	if isUpdateAvailable {
		m.newVersionAvailableMsg = fmt.Sprintf(
			"New version available: %s\nPlease visit: %s",
			version,
			m.releaseURL,
		)
	}
}

func (m *ModelInfo) testConnection(ctx context.Context) {
	defer m.skeleton.TriggerUpdate()

	m.status.SetProgressMessage("Checking your token...")
	m.skeleton.LockTabs()

	_, err := m.github.GetAuthUser(ctx)
	if err != nil {
		m.handleConnectionError(err)
		return
	}

	m.handleSuccessfulConnection()
}

// -----------------------------------------------------------------------------
// Error Handling
// -----------------------------------------------------------------------------

func (m *ModelInfo) handleUpdateError(err error) {
	m.status.SetError(err)
	m.status.SetErrorMessage("failed to check updates")
	m.newVersionAvailableMsg = fmt.Sprintf(
		"failed to check updates.\nPlease visit: %s",
		m.releaseURL,
	)
}

func (m *ModelInfo) handleConnectionError(err error) {
	m.status.SetError(err)
	m.status.SetErrorMessage("failed to test connection, please check your token&permission")
	m.skeleton.LockTabs()
}

func (m *ModelInfo) handleSuccessfulConnection() {
	m.status.Reset()
	m.status.SetSuccessMessage("Welcome to GAMA!")
	m.skeleton.UnlockTabs()
}
