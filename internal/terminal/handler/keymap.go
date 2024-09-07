package handler

import (
	"fmt"
	"github.com/termkit/gama/internal/config"

	teakey "github.com/charmbracelet/bubbles/key"
)

func loadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// ---------------------------------------------------------------------------

type handlerKeyMap struct {
	SwitchTabRight teakey.Binding
	SwitchTabLeft  teakey.Binding
	Quit           teakey.Binding
}

var handlerKeys = func() handlerKeyMap {
	cfg := loadConfig()

	return handlerKeyMap{
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

// ---------------------------------------------------------------------------

type githubInformationKeyMap struct {
	SwitchTabRight teakey.Binding
	Quit           teakey.Binding
}

func (k githubInformationKeyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTabRight, k.Quit}
}

func (k githubInformationKeyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTabRight},
		{k.Quit},
	}
}

var githubInformationKeys = func() githubInformationKeyMap {
	cfg := loadConfig()

	switchTabRight := cfg.Shortcuts.SwitchTabRight

	return githubInformationKeyMap{
		SwitchTabRight: teakey.NewBinding(
			teakey.WithKeys(""), // help-only binding
			teakey.WithHelp(switchTabRight, "next tab"),
		),
		Quit: teakey.NewBinding(
			teakey.WithKeys("q", cfg.Shortcuts.Quit),
			teakey.WithHelp("q", "quit"),
		),
	}
}()

func (m *ModelInfo) ViewHelp() string {
	return m.help.View(m.keys)
}

// ---------------------------------------------------------------------------

type githubRepositoryKeyMap struct {
	Refresh   teakey.Binding
	LaunchTab teakey.Binding
	SwitchTab teakey.Binding
}

func (k githubRepositoryKeyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTab, k.Refresh, k.LaunchTab}
}

func (k githubRepositoryKeyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTab},
		{k.Refresh},
		{k.LaunchTab},
	}
}

var githubRepositoryKeys = func() githubRepositoryKeyMap {
	cfg := loadConfig()

	var tabSwitch = fmt.Sprintf("%s | %s", cfg.Shortcuts.SwitchTabLeft, cfg.Shortcuts.SwitchTabRight)

	return githubRepositoryKeyMap{
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

func (m *ModelGithubRepository) ViewHelp() string {
	return m.help.View(m.Keys)
}

// ---------------------------------------------------------------------------

type githubWorkflowHistoryKeyMap struct {
	LaunchTab teakey.Binding
	Refresh   teakey.Binding
	SwitchTab teakey.Binding
	LiveMode  teakey.Binding
}

func (k githubWorkflowHistoryKeyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTab, k.Refresh, k.LaunchTab, k.LiveMode}
}

func (k githubWorkflowHistoryKeyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTab},
		{k.Refresh},
		{k.LaunchTab},
		{k.LiveMode},
	}
}

var githubWorkflowHistoryKeys = func() githubWorkflowHistoryKeyMap {
	cfg := loadConfig()

	var tabSwitch = fmt.Sprintf("%s | %s", cfg.Shortcuts.SwitchTabLeft, cfg.Shortcuts.SwitchTabRight)

	return githubWorkflowHistoryKeyMap{
		Refresh: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Refresh),
			teakey.WithHelp(cfg.Shortcuts.Refresh, "Refresh list"),
		),
		LaunchTab: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.Enter),
			teakey.WithHelp(cfg.Shortcuts.Enter, "Launch the selected option"),
		),
		LiveMode: teakey.NewBinding(
			teakey.WithKeys(cfg.Shortcuts.LiveMode),
			teakey.WithHelp(cfg.Shortcuts.LiveMode, "Toggle live mode"),
		),
		SwitchTab: teakey.NewBinding(
			teakey.WithKeys(""), // help-only binding
			teakey.WithHelp(tabSwitch, "switch tab"),
		),
	}
}()

func (m *ModelGithubWorkflowHistory) ViewHelp() string {
	return m.Help.View(m.keys)
}

// ---------------------------------------------------------------------------

type githubWorkflowKeyMap struct {
	SwitchTab teakey.Binding
}

func (k githubWorkflowKeyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTab}
}

func (k githubWorkflowKeyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTab},
	}
}

var githubWorkflowKeys = func() githubWorkflowKeyMap {
	cfg := loadConfig()

	var tabSwitch = fmt.Sprintf("%s | %s", cfg.Shortcuts.SwitchTabLeft, cfg.Shortcuts.SwitchTabRight)

	return githubWorkflowKeyMap{
		SwitchTab: teakey.NewBinding(
			teakey.WithKeys(""), // help-only binding
			teakey.WithHelp(tabSwitch, "switch tab"),
		),
	}
}()

func (m *ModelGithubWorkflow) ViewHelp() string {
	return m.help.View(m.keys)
}

// ---------------------------------------------------------------------------

type githubTriggerKeyMap struct {
	SwitchTabLeft teakey.Binding
	SwitchTab     teakey.Binding
	Trigger       teakey.Binding
	Refresh       teakey.Binding
}

func (k githubTriggerKeyMap) ShortHelp() []teakey.Binding {
	return []teakey.Binding{k.SwitchTabLeft, k.Refresh, k.SwitchTab, k.Trigger}
}

func (k githubTriggerKeyMap) FullHelp() [][]teakey.Binding {
	return [][]teakey.Binding{
		{k.SwitchTabLeft},
		{k.Refresh},
		{k.SwitchTab},
		{k.Trigger},
	}
}

var githubTriggerKeys = func() githubTriggerKeyMap {
	cfg := loadConfig()

	previousTab := cfg.Shortcuts.SwitchTabLeft

	return githubTriggerKeyMap{
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
	return m.help.View(m.Keys)
}
