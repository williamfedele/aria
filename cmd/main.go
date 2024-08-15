package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/williamfedele/chime/internal/audio"
)

type model struct {
	tracks       []audio.Track
	selected     int
	audioControl chan string
	audioFeed    chan string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.audioFeed <- m.tracks[m.selected].Path
			m.audioControl <- "play"
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.tracks)-1 {
				m.selected++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Select track to play:\n\n"
	for i, track := range m.tracks {
		if i == m.selected {
			s += fmt.Sprintf("> %s\n", track.Title)
		} else {
			s += fmt.Sprintf("  %s\n", track.Title)
		}
	}
	s += "\nPress enter to play, q to quit."
	return s
}

func main() {

	// Load music library
	// TODO: Allow user to specify their music library path through a config file
	tracks, err := audio.LoadLibrary("/Users/will/Music/library")
	if err != nil {
		panic(err)
	}

	m := model{tracks: tracks, selected: 0, audioControl: make(chan string), audioFeed: make(chan string)}
	go audio.PlayAudio(m.audioControl, m.audioFeed)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
