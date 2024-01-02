package ghtrigger

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	PreviousTab teakey.Binding
	Refresh     teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.Refresh}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.Refresh},
	}
}

var keys = keyMap{
	PreviousTab: teakey.NewBinding(
		teakey.WithKeys("shift+left"),
		teakey.WithHelp("shift + ←", "previous tab"),
	),
	Refresh: teakey.NewBinding(
		teakey.WithKeys("ctrl+r", "ctrl+R"),
		teakey.WithHelp("ctrl+r", "Refresh workflow"),
	),
}

func (m *ModelGithubTrigger) ViewHelp() string {
	return m.Help.View(m.Keys)
}
