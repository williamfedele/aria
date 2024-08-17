package ui

import (
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

	d := list.NewDefaultDelegate()
	help := []key.Binding{keys.TogglePlayback}
	d.ShortHelpFunc = func() []key.Binding {
		return help
	}
	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	library.Tracks.SetDelegate(d)

	library.Tracks.Styles.Title = titleStyle

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
		case key.Matches(msg, m.keys.TogglePlayback):
			track := m.Library.Tracks.SelectedItem().(audio.Track)
			m.Library.Tracks.NewStatusMessage(statusMessageStyle("> " + track.Title()))
			m.Player.Load(track)
			m.Player.Play()
		}
	}

	var cmd tea.Cmd
	m.Library.Tracks, cmd = m.Library.Tracks.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return appStyle.Render(m.Library.Tracks.View())
}
