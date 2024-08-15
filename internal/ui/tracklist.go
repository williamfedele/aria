package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamfedele/chime/internal/audio"
)

var style = lipgloss.NewStyle().Margin(1, 2)

type trackListModel struct {
	tracks       list.Model
	trackControl chan audio.Control
	trackFeed    chan string
}

func NewTrackListModel(tracks []audio.Track, trackControl chan audio.Control, trackFeed chan string) trackListModel {
	var items []list.Item
	for _, track := range tracks {
		items = append(items, track)
	}

	trackList := list.New(items, list.NewDefaultDelegate(), 0, 0)
	trackList.Title = "Chime"

	return trackListModel{
		tracks:       trackList,
		trackControl: trackControl,
		trackFeed:    trackFeed,
	}
}

func (m trackListModel) Init() tea.Cmd {
	return nil
}

func (m trackListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "p":
			m.trackControl <- audio.Pause
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if track, ok := m.tracks.SelectedItem().(audio.Track); ok {
				m.trackFeed <- track.Path
				m.trackControl <- audio.Play
			}
		}
	case tea.WindowSizeMsg:
		h, v := style.GetFrameSize()
		m.tracks.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.tracks, cmd = m.tracks.Update(msg)
	return m, cmd
}

func (m trackListModel) View() string {
	return style.Render(m.tracks.View())
}
