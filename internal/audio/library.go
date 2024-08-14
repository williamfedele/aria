package audio

import (
	"os"
	"path/filepath"
)

func LoadLibrary(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// TODO: Add support for nested directories
	// TODO: Add support for more audio formats
	// TODO: Add support for metadata displayed in the UI
	// TODO: Add support for album art
	// TODO: Display track title, not full file path

	var tracks []string
	for _, file := range files {
		if fileExt := filepath.Ext(file.Name()); fileExt != ".flac" {
			continue
		}

		tracks = append(tracks, filepath.Join(dir, file.Name()))
	}

	return tracks, nil
}
