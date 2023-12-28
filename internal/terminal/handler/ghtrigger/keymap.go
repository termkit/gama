package ghtrigger

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	PreviousTab teakey.Binding
	Refresh     teakey.Binding
	Quit        teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.Refresh, k.Quit}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.Refresh},
		{k.Quit},
	}
}

var keys = keyMap{
	PreviousTab: teakey.NewBinding(
		teakey.WithKeys("left"),
		teakey.WithHelp("←", "previous tab"),
	),
	Refresh: teakey.NewBinding(
		teakey.WithKeys("ctrl+r", "ctrl+R"),
		teakey.WithHelp("ctrl+r", "Refresh workflow"),
	),
	Quit: teakey.NewBinding(
		teakey.WithKeys("q", "ctrl+c"),
		teakey.WithHelp("q", "quit"),
	),
}

func (m *ModelGithubTrigger) ViewHelp() string {
	return m.Help.View(m.Keys)
}
