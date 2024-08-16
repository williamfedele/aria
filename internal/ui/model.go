package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamfedele/chime/internal/audio"
	"github.com/williamfedele/chime/internal/config"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(4)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("3"))
)

type Model struct {
	Player  *audio.Player
	Library *audio.Library
	Cursor  int
}

func InitialModel(config config.Config) Model {

	library, err := audio.NewLibrary(config.LibraryDir)
	if err != nil {
		panic(err)
	}

	return Model{
		Player:  audio.NewPlayer(),
		Library: library,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.Cursor < len(m.Library.Tracks)-1 {
				m.Cursor++
			}
		case "k", "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "enter":
			m.Player.Load(m.Library.Tracks[m.Cursor].Path)
			m.Player.Play()
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render("Tracks"))

	for i, track := range m.Library.Tracks {
		if i == m.Cursor {
			fmt.Fprintf(&b, "%s\n", selectedItemStyle.Render(track.Artist+" "+track.Title))
		} else {
			fmt.Fprintf(&b, "%s\n", itemStyle.Render(track.Artist+" "+track.Title))
		}
	}

	return b.String()
}
