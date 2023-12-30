package ghworkflow

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	TabSwitch teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.TabSwitch}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.TabSwitch},
	}
}

var keys = keyMap{
	TabSwitch: teakey.NewBinding(
		teakey.WithKeys("ctrl+left", "ctrl+right"),
		teakey.WithHelp("ctrl + (← | →)", "switch tab"),
	),
}

func (m *ModelGithubWorkflow) ViewHelp() string {
	return m.Help.View(m.Keys)
}
