package ghworkflow

import (
	"fmt"

	"github.com/termkit/gama/internal/config"

	teakey "github.com/charmbracelet/bubbles/key"
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
	cfg, err := config.LoadConfig()
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
