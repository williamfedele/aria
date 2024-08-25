package ui

import (
	"math/rand"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamfedele/aria/internal/audio"
	"github.com/williamfedele/aria/internal/config"
)

var (
	appStyle   = lipgloss.NewStyle().Padding(1, 2)
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#222222")).
			Background(lipgloss.Color("#FAD07B")).
			Padding(0, 2).Bold(true)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7A85FA")).
				Render
)

type Model struct {
	Player  *audio.Player
	Library *audio.Library
	keys    keyMap
}

func InitialModel(config config.Config) (Model, error) {

	library, err := audio.NewLibrary(config.LibraryDir)
	if err != nil {
		return Model{}, err
	}

	library.Tracks.Styles.Title = titleStyle
	library.Tracks.Title = "Nothing playing"

	m := Model{
		Player:  audio.NewPlayer(),
		Library: library,
		keys:    keys,
	}

	// Setup custom keybinds in the help section
	d := list.NewDefaultDelegate()
	d.ShortHelpFunc = func() []key.Binding {
		return []key.Binding{keys.Play}
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{{keys.Play, keys.TogglePlayback, keys.Stop, keys.Shuffle, keys.Enqueue}, {keys.Next, keys.Previous}, {keys.VolumeUp, keys.VolumeDown}}
	}

	d.Styles = list.DefaultItemStyles(NewDefaultItemStyles())
	library.Tracks.SetDelegate(d)

	return m, nil
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
			m.Player.ClearQueue()
			m.Player.ForcePlay(track)
			// TODO: Should the rest of the tracks be enqueues to autoplay?
		case key.Matches(msg, m.keys.Shuffle):
			// Reset the queue and stop playback for a fresh start
			m.Player.ClearQueue()
			m.Player.Stop()

			// Shuffle the tracks
			shuffledTracks := make([]list.Item, len(m.Library.Tracks.Items()))
			copy(shuffledTracks, m.Library.Tracks.Items())
			rand.Shuffle(len(shuffledTracks), func(i, j int) {
				shuffledTracks[i], shuffledTracks[j] = shuffledTracks[j], shuffledTracks[i]
			})
			var tracks []audio.Track
			for _, item := range shuffledTracks {
				tracks = append(tracks, item.(audio.Track))
			}
			m.Player.EnqueueAll(tracks)

		case key.Matches(msg, keys.TogglePlayback):
			m.Player.TogglePlayback()
		case key.Matches(msg, keys.Stop):
			m.Player.Stop()
		case key.Matches(msg, keys.Enqueue):
			track := m.Library.Tracks.SelectedItem().(audio.Track)
			m.Player.Enqueue(track)
		case key.Matches(msg, keys.Next):
			m.Player.Next()
		case key.Matches(msg, keys.Previous):
			m.Player.Previous()
		case key.Matches(msg, keys.VolumeUp):
			m.Player.VolumeUp()
		case key.Matches(msg, keys.VolumeDown):
			m.Player.VolumeDown()

		}
	case audio.PlaybackUpdate:
		if msg.CurrentTrack.Title() == "" {
			m.Library.Tracks.Title = "Nothing playing"
		} else {
			if msg.IsPlaying {
				m.Library.Tracks.Title = "Playing: " + msg.CurrentTrack.ShortString()
			} else {
				m.Library.Tracks.Title = "Paused: " + msg.CurrentTrack.ShortString()
			}
		}
		// Keep listening for playback updates
		return m, m.ListenForUpdates()
	case audio.StatusMessage:
		m.Library.Tracks.NewStatusMessage(statusMessageStyle(msg.Message))
		// Keep listening for playback updates
		return m, m.ListenForUpdates()
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
		select {
		case msg := <-m.Player.StatusMessage:
			return msg
		case msg := <-m.Player.PlaybackUpdate:
			return msg
		}
	}
}
