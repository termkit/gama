package ghworkflowhistory

import (
	"fmt"
	"github.com/termkit/gama/internal/config"

	teakey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	LaunchTab teakey.Binding
	Refresh   teakey.Binding
	SwitchTab teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTab, k.Refresh, k.LaunchTab}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTab},
		{k.Refresh},
		{k.LaunchTab},
	}
}

var keys = func() keyMap {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	var tabSwitch = fmt.Sprintf("%s | %s", cfg.Shortcuts.SwitchTabLeft, cfg.Shortcuts.SwitchTabRight)

	return keyMap{
		Refresh: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Refresh),
			teakey.WithHelp(cfg.Shortcuts.Refresh, "Refresh list"),
		),
		LaunchTab: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Enter),
			teakey.WithHelp(cfg.Shortcuts.Enter, "Launch the selected option"),
		),
		SwitchTab: teakey.NewBinding(
			teakey.WithKeys(""), // help-only binding
			teakey.WithHelp(tabSwitch, "switch tab"),
		),
	}
}()

func (m *ModelGithubWorkflowHistory) ViewHelp() string {
	return m.Help.View(m.Keys)
}
