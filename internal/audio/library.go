package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

type Library struct {
	Tracks list.Model
}

type Track struct {
	artist string
	album  string
	title  string
	path   string
	format AudioFormat
}

func (t Track) Artist() string      { return t.artist }
func (t Track) Album() string       { return t.album }
func (t Track) Title() string       { return t.title }
func (t Track) Path() string        { return t.path }
func (t Track) Format() AudioFormat { return t.format }
func (t Track) Description() string {
	return fmt.Sprintf("%s / %s / %s", t.Artist(), t.Album(), t.Title())
}
func (t Track) FilterValue() string {
	return t.Title()
}

func (t Track) String() string {
	return fmt.Sprintf("%s / %s / %s", t.Artist(), t.Album(), t.Title())
}

type AudioFormat int

const (
	FLAC AudioFormat = iota
	MP3
	WAV
	OGG
)

func NewLibrary(libraryDir string) (*Library, error) {

	// TODO: Add nested directory UI visual
	// TODO: Add support for metadata displayed in the UI
	// TODO: Add support for album art
	var tracks []Track
	err := filepath.WalkDir(libraryDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			track, err := extractTrack(libraryDir, path)
			if err != nil {
				return nil
			}
			tracks = append(tracks, track)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var items []list.Item
	for _, track := range tracks {
		items = append(items, track)
	}

	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Library"

	return &Library{Tracks: list}, nil

}

// Parses the path to extract the artist, album, title, and audio format
// The path should be in the format artist/album/title.format
// If the path is not in this format, the artist and album will be set to "unknown"
// The audio format is determined by the file extension
func extractTrack(rootDir string, fullPath string) (Track, error) {

	// Extract path relative to library root
	localPath := strings.TrimPrefix(fullPath, rootDir+"/")

	// There should be 3 levels: artist/album/track in that order
	levels := strings.Split(localPath, "/")

	// Extract audio format
	format := extractAudioFormat(localPath)
	if format == -1 {
		return Track{}, fmt.Errorf("unsupported format: %s", fullPath)
	}

	// If there are less than 3 dir levels, artist and album are set to "unknown" as a fallback
	if len(levels) < 3 {
		return Track{artist: "unknown", album: "unknown", title: levels[len(levels)-1], path: fullPath, format: format}, nil
	}

	// Otherwise, set artist and title as the library hierarchy dictates
	return Track{
		artist: levels[len(levels)-3],
		album:  levels[len(levels)-2],
		title:  levels[len(levels)-1],
		path:   fullPath,
		format: format,
	}, nil
}

// Determines the audio format based on the file extension
func extractAudioFormat(path string) AudioFormat {
	switch filepath.Ext(path) {
	case ".flac":
		return FLAC
	case ".mp3":
		return MP3
	case ".wav":
		return WAV
	case ".ogg":
		return OGG
	default:
		return -1
	}
}
