package config

func fillDefaultShortcuts(cfg *Config) *Config {
	var switchTabRight = cfg.Shortcuts.SwitchTabRight
	if switchTabRight == "" {
		switchTabRight = defaultKeyMap.SwitchTabRight
	}
	var switchTabLeft = cfg.Shortcuts.SwitchTabLeft
	if switchTabLeft == "" {
		switchTabLeft = defaultKeyMap.SwitchTabLeft
	}
	var quit = cfg.Shortcuts.Quit
	if quit == "" {
		quit = defaultKeyMap.Quit
	}
	var refresh = cfg.Shortcuts.Refresh
	if refresh == "" {
		refresh = defaultKeyMap.Refresh
	}
	var enter = cfg.Shortcuts.Enter
	if enter == "" {
		enter = defaultKeyMap.Enter
	}
	var tab = cfg.Shortcuts.Tab
	if tab == "" {
		tab = defaultKeyMap.Tab
	}
	var liveMode = cfg.Shortcuts.LiveMode
	if liveMode == "" {
		liveMode = defaultKeyMap.LiveMode
	}
	cfg.Shortcuts = Shortcuts{
		SwitchTabRight: switchTabRight,
		SwitchTabLeft:  switchTabLeft,
		Quit:           quit,
		Refresh:        refresh,
		LiveMode:       liveMode,
		Enter:          enter,
		Tab:            tab,
	}

	return cfg
}

type defaultMap struct {
	SwitchTabRight string
	SwitchTabLeft  string
	Quit           string
	Refresh        string
	Enter          string
	Tab            string
	LiveMode       string
}

var defaultKeyMap = defaultMap{
	SwitchTabRight: "shift+right",
	SwitchTabLeft:  "shift+left",
	Quit:           "ctrl+c",
	Refresh:        "ctrl+r",
	Enter:          "enter",
	Tab:            "tab",
	LiveMode:       "ctrl+l",
}
