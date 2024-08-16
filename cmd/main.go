package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamfedele/chime/internal/config"
	"github.com/williamfedele/chime/internal/ui"
)

func main() {

	// Load music library
	// TODO: Allow user to specify their music library path through a config file

	m := ui.InitialModel(config.NewConfig())

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
