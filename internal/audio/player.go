package audio

import (
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/speaker"
)

func PlayAudio(audioControl <-chan string, audioFeed <-chan string) error {

	var streamer beep.StreamSeekCloser
	var format beep.Format

	for {
		select {
		case path := <-audioFeed:

			if streamer != nil {
				speaker.Clear()
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			streamer, format, err = flac.Decode(f)
			if err != nil {
				return err
			}
			defer streamer.Close()

			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		case cmd := <-audioControl:
			switch cmd {
			case "play":
				speaker.Play(streamer)
			case "pause":
				speaker.Lock()
				speaker.Unlock()

			}
		}
	}
}
