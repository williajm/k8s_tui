package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the application
type KeyMap struct {
	// Global keys
	Quit    key.Binding
	Help    key.Binding
	Refresh key.Binding

	// Navigation keys
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Selection keys
	Enter key.Binding
	Back  key.Binding

	// View switching keys
	Tab      key.Binding
	ShiftTab key.Binding

	// Resource actions
	Namespace key.Binding
	Context   key.Binding
	Search    key.Binding
	Logs      key.Binding
	Events    key.Binding
	YAML      key.Binding
	Describe  key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		// Global keys
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "f5"),
			key.WithHelp("r", "refresh"),
		),

		// Navigation keys
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "bottom"),
		),

		// Selection keys
		Enter: key.NewBinding(
			key.WithKeys("enter", "right", "l"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("left", "h", "backspace", "esc"),
			key.WithHelp("esc", "back"),
		),

		// View switching
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev pane"),
		),

		// Resource actions
		Namespace: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "namespace"),
		),
		Context: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "context"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Logs: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "logs"),
		),
		Events: key.NewBinding(
			key.WithKeys("E"),
			key.WithHelp("E", "events"),
		),
		YAML: key.NewBinding(
			key.WithKeys("Y"),
			key.WithHelp("Y", "yaml"),
		),
		Describe: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "describe"),
		),
	}
}

// ShortHelp returns a slice of key bindings to be displayed in short help
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Enter,
		k.Search,
		k.Refresh,
		k.Help,
		k.Quit,
	}
}

// FullHelp returns all key bindings organized by category
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		// Navigation
		{k.Up, k.Down, k.PageUp, k.PageDown, k.Home, k.End},
		// Selection
		{k.Enter, k.Back, k.Tab, k.ShiftTab},
		// Actions
		{k.Namespace, k.Context, k.Search, k.Refresh},
		// Resource actions
		{k.Logs, k.Events, k.YAML, k.Describe},
		// Global
		{k.Help, k.Quit},
	}
}
