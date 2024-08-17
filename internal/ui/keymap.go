package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	TogglePlayback key.Binding
}

var keys = keyMap{
	TogglePlayback: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "toggle playback"),
	),
}
