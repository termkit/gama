package ghworkflowhistory

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Sort        teakey.Binding
	NextTab     teakey.Binding
	PreviousTab teakey.Binding
	Quit        teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.NextTab, k.Sort, k.Quit}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.NextTab},
		{k.Sort},
		{k.Quit},
	}
}

var keys = keyMap{
	Sort: teakey.NewBinding(
		teakey.WithKeys("s", "S"),
		teakey.WithHelp("s/S", "Sort By Time"),
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

func (m *ModelGithubWorkflowHistory) ViewHelp() string {
	return m.Help.View(m.Keys)
}
