package common

import "github.com/charmbracelet/bubbles/key"

type HelpBindable interface {
	HelpBindings() []key.Binding
}

type FullHelpBindable interface {
	FullHelpBindings() []key.Binding
}

type KeyMap struct {
	Help   key.Binding
	Back   key.Binding
	Quit   key.Binding
	Logout key.Binding

	Library key.Binding
	Search  key.Binding
	Lists   key.Binding
	Stats   key.Binding
	NextTab key.Binding
	PrevTab key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextTab, k.PrevTab},
		{k.Help, k.Back, k.Logout, k.Quit},
	}
}

var Keys = KeyMap{
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Logout: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "logout"),
	),
	Library: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "home"),
	),
	Search: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "search"),
	),
	Lists: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "lists"),
	),
	Stats: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "stats"),
	),
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
		key.WithDisabled(),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
		key.WithDisabled(),
	),
}
