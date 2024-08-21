package main

import (
	"flag"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamfedele/chime/internal/config"
	"github.com/williamfedele/chime/internal/ui"
)

func main() {

	// Load music library
	// TODO: Allow user to specify their music library path through a config file
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Please provide a path to your music library")
		return
	}

	m, err := ui.InitialModel(config.NewConfig(args[0]))
	if err != nil {
		fmt.Println("Error initializing model:", err)
		return
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
