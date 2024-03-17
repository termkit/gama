package handler

import (
	"fmt"

	teakey "github.com/charmbracelet/bubbles/key"
	pkgconfig "github.com/termkit/gama/pkg/config"
)

type keyMap struct {
	SwitchTabRight teakey.Binding
	SwitchTabLeft  teakey.Binding
	Quit           teakey.Binding
}

var keys = func() keyMap {
	cfg, err := pkgconfig.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return keyMap{
		SwitchTabRight: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.SwitchTabRight),
		),
		SwitchTabLeft: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.SwitchTabLeft),
		),
		Quit: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Quit),
		),
	}
}()
