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

	trackControl := make(chan audio.Control)
	trackFeed := make(chan string)

	m := ui.NewTrackListModel(tracks, trackControl, trackFeed)
	go audio.PlayAudio(trackControl, trackFeed)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
