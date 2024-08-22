package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Play           key.Binding
	Shuffle        key.Binding
	TogglePlayback key.Binding
	Stop           key.Binding
	Enqueue        key.Binding
	Skip           key.Binding
	VolumeUp       key.Binding
	VolumeDown     key.Binding
}

var keys = keyMap{
	Play: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "play"),
	),
	TogglePlayback: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle playback"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop playback"),
	),
	Shuffle: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "shuffle"),
	),
	Enqueue: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add to queue"),
	),
	Skip: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "skip track"),
	),
	VolumeUp: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "volume up"),
	),
	VolumeDown: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "volume down"),
	),
}
