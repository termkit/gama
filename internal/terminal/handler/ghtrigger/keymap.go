package ghtrigger

import (
	"fmt"
	"github.com/termkit/gama/internal/config"

	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	SwitchTabLeft teakey.Binding
	SwitchTab     teakey.Binding
	Trigger       teakey.Binding
	Refresh       teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTabLeft, k.Refresh, k.SwitchTab, k.Trigger}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTabLeft},
		{k.Refresh},
		{k.SwitchTab},
		{k.Trigger},
	}
}

var keys = func() keyMap {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	var previousTab = fmt.Sprintf("%s", cfg.Shortcuts.SwitchTabLeft)

	return keyMap{
		SwitchTabLeft: teakey.NewBinding(
			teakey.WithKeys(""), // help-only binding
			teakey.WithHelp(previousTab, "previous tab"),
		),
		Refresh: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Refresh),
			teakey.WithHelp(cfg.Shortcuts.Refresh, "Refresh list"),
		),
		SwitchTab: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Tab),
			teakey.WithHelp(cfg.Shortcuts.Tab, "switch button"),
		),
		Trigger: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Enter),
			teakey.WithHelp(cfg.Shortcuts.Enter, "trigger workflow"),
		),
	}
}()

func (m *ModelGithubTrigger) ViewHelp() string {
	return m.Help.View(m.Keys)
}
