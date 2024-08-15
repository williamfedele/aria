package audio

import (
	"os"
	"path/filepath"
)

type Track struct {
	Title  string
	Path   string
	Format string
}

func LoadLibrary(libraryDir string) ([]Track, error) {
	files, err := os.ReadDir(libraryDir)
	if err != nil {
		return nil, err
	}

	// TODO: Add nested directory UI visual
	// TODO: Add support for more audio formats
	// TODO: Add support for metadata displayed in the UI
	// TODO: Add support for album art

	var tracks []Track
	for _, file := range files {

		fileInfo, err := file.Info()
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			err := filepath.WalkDir(filepath.Join(libraryDir, file.Name()), func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if fileExt := filepath.Ext(d.Name()); fileExt != ".flac" {
					return nil
				}
				relativePath, err := filepath.Rel(libraryDir, path)
				if err != nil {
					return err
				}
				tracks = append(tracks, Track{Title: d.Name(), Path: filepath.Join(libraryDir, relativePath), Format: "flac"})
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			if fileExt := filepath.Ext(file.Name()); fileExt != ".flac" {
				continue
			}
			tracks = append(tracks, Track{Title: file.Name(), Path: filepath.Join(libraryDir, file.Name()), Format: "flac"})
		}
	}

	return tracks, nil
}
