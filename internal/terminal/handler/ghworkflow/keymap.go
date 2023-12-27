package ghworkflow

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	NextTab     teakey.Binding
	PreviousTab teakey.Binding
	Quit        teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.NextTab, k.Quit}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.NextTab},
		{k.Quit},
	}
}

var keys = keyMap{
	PreviousTab: teakey.NewBinding(
		teakey.WithKeys("left"),
		teakey.WithHelp("←", "previous tab"),
	),
	NextTab: teakey.NewBinding(
		teakey.WithKeys("right"),
		teakey.WithHelp("→", "trigger page"),
	),
	Quit: teakey.NewBinding(
		teakey.WithKeys("q", "ctrl+c"),
		teakey.WithHelp("q", "quit"),
	),
}

func (m *ModelGithubWorkflow) ViewHelp() string {
	return m.Help.View(m.Keys)
}
