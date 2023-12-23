package ghrepository

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Search      teakey.Binding
	NextTab     teakey.Binding
	PreviousTab teakey.Binding
	Quit        teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.NextTab, k.Search, k.Quit}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.NextTab},
		{k.Search},
		{k.Quit},
	}
}

var keys = keyMap{
	Search: teakey.NewBinding(
		teakey.WithKeys("/"),
		teakey.WithHelp("/", "Search repository"),
	),
	PreviousTab: teakey.NewBinding(
		teakey.WithKeys("left"),
		teakey.WithHelp("←", "previous tab"),
	),
	NextTab: teakey.NewBinding(
		teakey.WithKeys("right"),
		teakey.WithHelp("→", "next tab"),
	),
	Quit: teakey.NewBinding(
		teakey.WithKeys("q", "ctrl+c"),
		teakey.WithHelp("q", "quit"),
	),
}

func (m *ModelGithubRepository) ViewHelp() string {
	return m.Help.View(m.Keys)
}
