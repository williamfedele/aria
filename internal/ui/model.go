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
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.Library.Tracks.SetSize(msg.Width-h-1, msg.Height-v-1)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Play):
			track := m.Library.Tracks.SelectedItem().(audio.Track)
			m.Library.Tracks.NewStatusMessage(statusMessageStyle("Playing: " + track.Title()))
			m.Player.Load(track)
			m.Player.Play()
		case key.Matches(msg, m.keys.Shuffle):
			track := m.Library.Tracks.Items()[rand.Intn(len(m.Library.Tracks.Items()))].(audio.Track)
			m.Library.Tracks.NewStatusMessage(statusMessageStyle("Playing: " + track.Title()))
			m.Player.Load(track)
			m.Player.Play()
		case key.Matches(msg, keys.TogglePlayback):
			// TODO: need to receive event from the player in order to update the status message here
			m.Player.TogglePlayback()
		case key.Matches(msg, keys.Stop):
			m.Library.Tracks.NewStatusMessage(statusMessageStyle("Nothing playing"))
			m.Player.Stop()
		}

	}

	var cmd tea.Cmd
	m.Library.Tracks, cmd = m.Library.Tracks.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return appStyle.Render(m.Library.Tracks.View())
}
