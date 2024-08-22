<br/><br/>
<div>
    <h3 align="center">🎧 Aria</h3>
    <p align="center">
        A simple command line player for your digitally downloaded music. 
    </p>
    <p align="center">
        <img src="assets/preview.png">
    </p>
</div>


## Getting Started

### Prerequisites

Install [Go](https://go.dev) if it's not already installed.

### Installation
```sh
go install github.com/williamfedele/aria/cmd/aria@latest
```

## Usage

Until a proper configuration system is in place, the root directory of your music library must be provided as an argument.
```sh
aria /path/to/your/music/library/root
```

The application recursively searches the root directory for all audio files with the format `{TRACK_NAME}.{flac|mp3|wav|ogg}`. 

Library directory structure is defined as follows:

```
ROOT_DIR
  ARTIST_NAME
    ALBUM1_NAME
      TRACK1_NAME.{flac|mp3|wav|ogg}
      TRACK2_NAME.{flac|mp3|wav|ogg}
    ALBUM2_NAME
      TRACK1_NAME.{flac|mp3|wav|ogg}
```

If an audio file is found that does not follow this structure, it will still be playable but will have the artist and album set to `unknown`.

### Keybinds

Keybinds can be found in the help view at the bottom of the terminal. It can be expanded using `?` to show the full keybind list.

Apart from the Vim-like movement, the following keybinds are set for manipulating the music player:

- `enter`: Plays the currently highlighted track.
- `space`: Toggles the playback between play/pause.
- `x`: Stops all playback and clears the queue.
- `S`: Shuffles all tracks and plays the first.
- `a`: Adds the currently highlighted track to the queue.
- `s`: Skip the currently playing track.


## Development

This was a big learning experience about concurrency in Go and it's likely got bugs. I'm still working on this so I'll keep hunting them down as I add new features. 

I'd like to add the following soon:

- Proper configuration system
- More controls over the playback such as volume


## License

[MIT](https://github.com/williamfedele/aria/blob/main/LICENSE)


## Acknowledgments

* [Charmbracelet](https://github.com/charmbracelet) for their incredibly cool CLI tools.
