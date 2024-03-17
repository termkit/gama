package ghworkflow

import (
	"fmt"

	teakey "github.com/charmbracelet/bubbles/key"
	pkgconfig "github.com/termkit/gama/pkg/config"
)

type keyMap struct {
	SwitchTab teakey.Binding
}

func (k keyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTab}
}

func (k keyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTab},
	}
}

var keys = func() keyMap {
	cfg, err := pkgconfig.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	var tabSwitch = fmt.Sprintf("%s | %s", cfg.Shortcuts.SwitchTabLeft, cfg.Shortcuts.SwitchTabRight)

	return keyMap{
		SwitchTab: teakey.NewBinding(
			teakey.WithKeys(""), // help-only binding
			teakey.WithHelp(tabSwitch, "switch tab"),
		),
	}
}()

func (m *ModelGithubWorkflow) ViewHelp() string {
	return m.Help.View(m.Keys)
}
