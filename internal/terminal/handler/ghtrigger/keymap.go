package ghtrigger

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	PreviousTab teakey.Binding
	SwitchTab   teakey.Binding
	Trigger     teakey.Binding
	Refresh     teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.PreviousTab, k.Refresh, k.SwitchTab, k.Trigger}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.PreviousTab},
		{k.Refresh},
		{k.SwitchTab},
		{k.Trigger},
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
	SwitchTab: teakey.NewBinding(
		teakey.WithKeys("tab"),
		teakey.WithHelp("tab", "switch between tabs"),
	),
	Trigger: teakey.NewBinding(
		teakey.WithKeys("enter"),
		teakey.WithHelp("enter", "trigger workflow"),
	),
}

func (m *ModelGithubTrigger) ViewHelp() string {
	return m.Help.View(m.Keys)
}
