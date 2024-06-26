package header

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"sync"
)

// Header is a helper for rendering the Header of the terminal.
type Header struct {
	keys keyMap

	Viewport *viewport.Model
	//timer    timer.Model

	currentTab int
	lockTabs   bool

	commonHeaders  []commonHeader
	specialHeaders []specialHeader

	switchSpecialAnimation bool
}

type commonHeader struct {
	header    string
	rawHeader string

	inactiveStyle lipgloss.Style
	activeStyle   lipgloss.Style
}

type specialHeader struct {
	header    string
	rawHeader string

	firstStyle  lipgloss.Style
	secondStyle lipgloss.Style
}

// Define sync.Once and NewHeader should return same instance
var (
	once sync.Once
	h    *Header
)

// NewHeader returns a new Header.
func NewHeader(viewport *viewport.Model) *Header {
	once.Do(func() {
		h = &Header{
			Viewport:   viewport,
			currentTab: 0,
			lockTabs:   true,
			keys:       keys,
		}
	})
	return h
}

func (h *Header) SetCurrentTab(tab int) {
	h.currentTab = tab
}

func (h *Header) GetCurrentTab() int {
	return h.currentTab
}

func (h *Header) SetLockTabs(lock bool) {
	h.lockTabs = lock
}

func (h *Header) GetLockTabs() bool {
	return h.lockTabs
}

func (h *Header) AddCommonHeader(header string, inactiveStyle, activeStyle lipgloss.Style) {
	h.commonHeaders = append(h.commonHeaders, commonHeader{
		header:        header,
		rawHeader:     header,
		inactiveStyle: inactiveStyle,
		activeStyle:   activeStyle,
	})
}

func (h *Header) SetSpecialHeader(header string, firstStyle, secondStyle lipgloss.Style) {
	h.specialHeaders = append(h.specialHeaders, specialHeader{
		header:      header,
		rawHeader:   header,
		firstStyle:  firstStyle,
		secondStyle: secondStyle,
	})
}

func (h *Header) Init() tea.Cmd {
	//return h.timer.Init()
	return nil
}

func (h *Header) Update(msg tea.Msg) (*Header, tea.Cmd) {
	//var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keys.SwitchTabLeft):
			if !h.lockTabs {
				h.currentTab = max(h.currentTab-1, 0)
			}
		case key.Matches(msg, h.keys.SwitchTabRight):
			if !h.lockTabs {
				h.currentTab = min(h.currentTab+1, len(h.commonHeaders)-1)
			}
		}

		//case timer.TickMsg:
		//h.switchSpecialAnimation = !h.switchSpecialAnimation

		//var cmd tea.Cmd
		//h.timer, cmd = h.timer.Update(msg)
		//return h, cmd
		//
		//case timer.StartStopMsg:
		//	var cmd tea.Cmd
		//h.timer, cmd = h.timer.Update(msg)
		//return h, cmd

		//case tea.KeyMsg:
		//	switch {
		//	case key.Matches(msg, m.keymap.quit):
		//		m.quitting = true
		//		return m, tea.Quit
		//	case key.Matches(msg, m.keymap.reset):
		//		m.timer.Timeout = timeout
		//	case key.Matches(msg, m.keymap.start, m.keymap.stop):
		//		return m, m.timer.Toggle()
		//	}
	}

	return h, nil
}

// View renders the Header.
func (h *Header) View() string {
	var titles string
	titles += "BBEEE"
	for _, title := range h.commonHeaders {
		titles += title.rawHeader
		titles += "LLL RRR"
	}
	titleLen := len(titles)
	specialTitleLen := len(h.specialHeaders[0].header)

	var specialHeader string
	specialHeader = h.specialHeaders[0].header

	var renderedTitles []string
	for i, title := range h.commonHeaders {
		if i == h.currentTab {
			renderedTitles = append(renderedTitles, title.activeStyle.Render(title.header))
		} else {
			renderedTitles = append(renderedTitles, title.inactiveStyle.Render(title.header))
		}
	}

	if h.currentTab == -1 {
		specialHeader = h.specialHeaders[0].firstStyle.Render(h.specialHeaders[0].header)
	} else {
		specialHeader = h.specialHeaders[0].secondStyle.Render(h.specialHeaders[0].header)
	}

	line := strings.Repeat("─", h.Viewport.Width-(titleLen+specialTitleLen))

	return lipgloss.JoinHorizontal(lipgloss.Center, append(renderedTitles, line, specialHeader)...)
}
