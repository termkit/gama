package skeleton

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/termkit/gama/internal/terminal/handler/header"
	"github.com/termkit/gama/internal/terminal/handler/spirit"
	"github.com/termkit/gama/internal/terminal/handler/types"
	"sync"
)

// Skeleton is a helper for rendering the Skeleton of the terminal.
type Skeleton struct {
	Viewport *viewport.Model

	header      *header.Header
	modelSpirit *spirit.ModelSpirit
	lockTabs    bool

	keys keyMap

	currentTab int
	Pages      []tea.Model
}

func (s *Skeleton) AddPage(title Title, page tea.Model) {
	s.header.AddCommonHeader(title.Title, title.Style.Active, title.Style.Inactive)
	s.Pages = append(s.Pages, page)
}

type Title struct {
	Title string
	Style TitleStyle
}

type TitleStyle struct {
	Active   lipgloss.Style
	Inactive lipgloss.Style
}

var (
	once sync.Once
	s    *Skeleton
)

// NewSkeleton returns a new Skeleton.
func NewSkeleton() *Skeleton {
	once.Do(func() {
		s = &Skeleton{
			Viewport:    types.NewTerminalViewport(),
			header:      header.NewHeader(),
			modelSpirit: spirit.NewSpirit(),
			keys:        keys,
		}
	})
	return s
}

type SwitchTab struct {
	Tab int
}

func (s *Skeleton) SetCurrentTab(tab int) {
	s.currentTab = tab
}

func (s *Skeleton) Init() tea.Cmd {
	self := func() tea.Msg {
		return SwitchTab{}
	}

	inits := make([]tea.Cmd, len(s.Pages)+1) // +1 for self
	for i := range s.Pages {
		inits[i] = s.Pages[i].Init()
	}

	inits[len(s.Pages)] = self

	return tea.Batch(inits...)
}

func (s *Skeleton) Update(msg tea.Msg) (*Skeleton, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	s.currentTab = s.header.GetCurrentTab()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Viewport.Width = msg.Width
		s.Viewport.Height = msg.Height
	//case SwitchTab:
	//	s.SetCurrentTab(msg.Tab)
	//	s.header.SetCurrentTab(msg.Tab)
	//
	//	var cmd tea.Cmd
	//	s.Pages[s.currentTab], cmd = s.Pages[msg.Tab].Update(msg)
	//	return s, cmd
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keys.Quit):
			return s, tea.Quit
		case key.Matches(msg, s.keys.SwitchTabLeft):
			if !s.modelSpirit.GetLockTabs() {
				s.currentTab = max(s.currentTab-1, 0)
			}
		case key.Matches(msg, s.keys.SwitchTabRight):
			if !s.modelSpirit.GetLockTabs() {
				s.currentTab = min(s.currentTab+1, len(s.Pages)-1)
			}
		}
	}

	s.header, cmd = s.header.Update(msg)
	cmds = append(cmds, cmd)

	s.Pages[s.currentTab], cmd = s.Pages[s.currentTab].Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s *Skeleton) View() string {
	return lipgloss.JoinVertical(lipgloss.Top, s.header.View(), s.Pages[s.currentTab].View())
}
