package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamfedele/chime/internal/audio"
)

var style = lipgloss.NewStyle().Margin(1, 2)

type Model struct {
	List         list.Model
	TrackControl chan string
	TrackFeed    chan string
}

func NewModel(tracks []audio.Track) Model {
	var items []list.Item
	for _, track := range tracks {
		items = append(items, track)
	}

	return Model{
		List:         list.New(items, list.NewDefaultDelegate(), 0, 0),
		TrackControl: make(chan string),
		TrackFeed:    make(chan string),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if track, ok := m.List.SelectedItem().(audio.Track); ok {
				m.TrackFeed <- track.Path
				m.TrackControl <- "play"
			}
		}
	case tea.WindowSizeMsg:
		h, v := style.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return style.Render(m.List.View())
}
