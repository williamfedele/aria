package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamfedele/chime/internal/audio"
	"github.com/williamfedele/chime/internal/ui"
)

func main() {

	// Load music library
	// TODO: Allow user to specify their music library path through a config file
	tracks, err := audio.LoadLibrary("/Users/will/Music/library")
	if err != nil {
		panic(err)
	}

	m := ui.NewModel(tracks)
	m.List.Title = "Music Library"
	go audio.PlayAudio(m.TrackControl, m.TrackFeed)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
