package ui

import (
	"math/rand"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamfedele/chime/internal/audio"
	"github.com/williamfedele/chime/internal/config"
)

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#222222")).
			Background(lipgloss.Color("3")).
			Padding(0, 2)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("3")).
				Render
)

type Model struct {
	Player  *audio.Player
	Library *audio.Library
	keys    keyMap
}

func InitialModel(config config.Config) Model {

	library, err := audio.NewLibrary(config.LibraryDir)
	if err != nil {
		panic(err)
	}

	m := Model{
		Player:  audio.NewPlayer(),
		Library: library,
		keys:    keys,
	}

	// Setup custom keybinds
	d := list.NewDefaultDelegate()

	d.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{keys.Play}
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{keys.Play, keys.TogglePlayback, keys.Stop, keys.Shuffle}}
	}

	library.Tracks.SetDelegate(d)

	library.Tracks.Styles.Title = titleStyle
	m.Library.Tracks.NewStatusMessage(statusMessageStyle("Nothing playing"))

	return m
}

func (m Model) Init() tea.Cmd {
	return m.ListenForUpdates()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.Library.Tracks.SetSize(msg.Width-h-1, msg.Height-v-1)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Play):
			// Explicit track playing will clear the queue and immediately play
			track := m.Library.Tracks.SelectedItem().(audio.Track)
			m.Player.ForcePlay(track)
		case key.Matches(msg, m.keys.Shuffle):
			// TODO: shuffling a second time should reset the queue
			// TODO: shuffle all tracks and enqueue them
			//track := m.Library.Tracks.Items()[rand.Intn(len(m.Library.Tracks.Items()))].(audio.Track)
			track := m.Library.Tracks.Items()[rand.Intn(len(m.Library.Tracks.Items()))].(audio.Track)
			track2 := m.Library.Tracks.Items()[rand.Intn(len(m.Library.Tracks.Items()))].(audio.Track)
			//m.Library.Tracks.NewStatusMessage(statusMessageStyle("Playing: " + track.Title()))
			m.Player.Enqueue(track)
			m.Player.Enqueue(track2)
		case key.Matches(msg, keys.TogglePlayback):
			m.Player.TogglePlayback()
		case key.Matches(msg, keys.Stop):
			m.Player.Stop()
		case key.Matches(msg, keys.Enqueue):
			track := m.Library.Tracks.SelectedItem().(audio.Track)
			m.Player.Enqueue(track)
		case key.Matches(msg, keys.Skip):
			m.Player.Next()

		}
	case audio.PlaybackUpdate:
		if msg.CurrentTrack.Title() == "" {
			m.Library.Tracks.Title = "Nothing playing"
		} else {
			if msg.IsPlaying {
				m.Library.Tracks.Title = "Playing: " + msg.CurrentTrack.Title()
			} else {
				m.Library.Tracks.Title = "Paused: " + msg.CurrentTrack.Title()
			}
		}
		// Keep listening for playback updates
		return m, m.ListenForUpdates()
	case audio.StatusMessage:
		m.Library.Tracks.NewStatusMessage(statusMessageStyle(msg.Message))
	}

	var cmd tea.Cmd
	m.Library.Tracks, cmd = m.Library.Tracks.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return appStyle.Render(m.Library.Tracks.View())
}

// Waits for messages from the player about any type of update
func (m Model) ListenForUpdates() tea.Cmd {
	return func() tea.Msg {
		return <-m.Player.PlaybackUpdate
		// select {
		// case msg := <-m.Player.StatusMessage:
		// 	return msg
		// case msg := <-m.Player.PlaybackUpdate:
		// 	return msg
		// }
	}
}
