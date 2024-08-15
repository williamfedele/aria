package main

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamfedele/chime/internal/audio"
)

var style = lipgloss.NewStyle().Margin(1, 2)

type model struct {
	list         list.Model
	trackControl chan string
	trackFeed    chan string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if track, ok := m.list.SelectedItem().(audio.Track); ok {
				m.trackFeed <- track.Path
				m.trackControl <- "play"
			}
		}
	case tea.WindowSizeMsg:
		h, v := style.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return style.Render(m.list.View())
}

func main() {

	// Load music library
	// TODO: Allow user to specify their music library path through a config file
	tracks, err := audio.LoadLibrary("/Users/will/Music/library")
	if err != nil {
		panic(err)
	}

	var items []list.Item
	for _, track := range tracks {
		items = append(items, track)
	}

	m := model{
		list:         list.New(items, list.NewDefaultDelegate(), 0, 0),
		trackControl: make(chan string),
		trackFeed:    make(chan string),
	}
	m.list.Title = "Music Library"
	go audio.PlayAudio(m.trackControl, m.trackFeed)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
