package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Play           key.Binding
	Shuffle        key.Binding
	TogglePlayback key.Binding
	Stop           key.Binding
}

var keys = keyMap{
	Play: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "play"),
	),
	Shuffle: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "shuffle"),
	),
	TogglePlayback: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "toggle playback"),
	),
	Stop: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "stop playback"),
	),
}
