package ghworkflowhistory

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	LaunchTab teakey.Binding
	Refresh   teakey.Binding
	TabSwitch teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.TabSwitch, k.Refresh, k.LaunchTab}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.TabSwitch},
		{k.Refresh},
		{k.LaunchTab},
	}
}

var keys = keyMap{
	Refresh: teakey.NewBinding(
		teakey.WithKeys("r", "R"),
		teakey.WithHelp("r/R", "Refresh list"),
	),
	LaunchTab: teakey.NewBinding(
		teakey.WithKeys("enter"),
		teakey.WithHelp("enter", "Launch the selected option"),
	),
	TabSwitch: teakey.NewBinding(
		teakey.WithKeys("shift+left", "shift+right"),
		teakey.WithHelp("shift + (← | →)", "switch tab"),
	),
}

func (m *ModelGithubWorkflowHistory) ViewHelp() string {
	return m.Help.View(m.Keys)
}
