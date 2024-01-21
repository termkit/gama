package ghrepository

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Refresh      teakey.Binding
	LaunchOption teakey.Binding
	TabSwitch    teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.TabSwitch, k.Refresh, k.LaunchOption}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.TabSwitch},
		{k.Refresh},
		{k.LaunchOption},
	}
}

var keys = keyMap{
	Refresh: teakey.NewBinding(
		teakey.WithKeys("r", "R"),
		teakey.WithHelp("r/R", "Refresh list"),
	),
	LaunchOption: teakey.NewBinding(
		teakey.WithKeys("enter"),
		teakey.WithHelp("enter", "Launch the selected option"),
	),
	TabSwitch: teakey.NewBinding(
		teakey.WithKeys(""), // help-only binding
		teakey.WithHelp("shift + (← | →)", "switch tab"),
	),
}

func (m *ModelGithubRepository) ViewHelp() string {
	return m.Help.View(m.Keys)
}
