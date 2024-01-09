package handler

import (
	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	SwitchTabRight teakey.Binding
	SwitchTabLeft  teakey.Binding
	Quit           teakey.Binding
}

var keys = keyMap{
	SwitchTabRight: teakey.NewBinding(
		teakey.WithKeys("shift+right"),
	),
	SwitchTabLeft: teakey.NewBinding(
		teakey.WithKeys("shift+left"),
	),
	Quit: teakey.NewBinding(
		teakey.WithKeys("ctrl+q"),
	),
}
