package header

import (
	"fmt"
	"github.com/termkit/gama/internal/config"

	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	SwitchTabRight teakey.Binding
	SwitchTabLeft  teakey.Binding
	Quit           teakey.Binding
}

var keys = func() keyMap {
	cfg, err := config.LoadConfig()
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
