package audio

import (
	"os"
	"time"

	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/speaker"
)

func PlayAudio(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	streamer, format, err := flac.Decode(f)
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(streamer)
	// TODO figure out how to make playing non-blocking
	select {}

	return nil
}
