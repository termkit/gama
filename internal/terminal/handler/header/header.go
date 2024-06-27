package header

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ts "github.com/termkit/gama/internal/terminal/style"
	"strings"
	"sync"
	"time"
)

// Header is a helper for rendering the Header of the terminal.
type Header struct {
	keys keyMap

	Viewport *viewport.Model

	tickerInterval time.Duration

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
			tickerInterval: time.Millisecond * 250,
			Viewport:       viewport,
			currentTab:     0,
			lockTabs:       true,
			keys:           keys,
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

type UpdateMsg struct {
	Msg               string
	UpdatingComponent string
}

func (h *Header) Init() tea.Cmd {
	return h.tick()
}

func (h *Header) tick() tea.Cmd {
	t := time.NewTimer(h.tickerInterval)
	return func() tea.Msg {
		//ts := <-t.C
		//t.Stop()
		//for len(t.C) > 0 {
		//	<-t.C
		//}
		//return UpdateMsg{
		//	Msg:               "tick",
		//	UpdatingComponent: "header",
		//}

		select {
		case <-t.C:
			return UpdateMsg{
				Msg:               "tick",
				UpdatingComponent: "header",
			}
		}
	}
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
	case UpdateMsg:
		h.switchSpecialAnimation = !h.switchSpecialAnimation

		return h, h.Init()
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
		if h.lockTabs {
			if i == 0 {
				renderedTitles = append(renderedTitles, title.activeStyle.Render(title.header))
			} else {
				renderedTitles = append(renderedTitles, ts.TitleStyleDisabled.Render(title.header))
			}
		} else {
			if i == h.currentTab {
				renderedTitles = append(renderedTitles, title.activeStyle.Render(title.header))
			} else {
				renderedTitles = append(renderedTitles, title.inactiveStyle.Render(title.header))
			}
		}
	}

	if h.switchSpecialAnimation {
		specialHeader = h.specialHeaders[0].firstStyle.Render(h.specialHeaders[0].header)
	} else {
		specialHeader = h.specialHeaders[0].secondStyle.Render(h.specialHeaders[0].header)
	}

	line := strings.Repeat("─", h.Viewport.Width-(titleLen+specialTitleLen))

	return lipgloss.JoinHorizontal(lipgloss.Center, append(renderedTitles, line, specialHeader)...)
}
