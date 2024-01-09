package information

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	NextTab teakey.Binding
	Quit    teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.NextTab, k.Quit}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.NextTab},
		{k.Quit},
	}
}

var keys = keyMap{
	NextTab: teakey.NewBinding(
		teakey.WithKeys(""), // help-only binding
		teakey.WithHelp("shift + →", "next tab"),
	),
	Quit: teakey.NewBinding(
		teakey.WithKeys("q", "ctrl+c"),
		teakey.WithHelp("q", "quit"),
	),
}

func (m *ModelInfo) ViewHelp() string {
	return m.Help.View(m.Keys)
}
