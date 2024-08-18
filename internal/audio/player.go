package audio

import (
	"fmt"
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/vorbis"
	"github.com/gopxl/beep/v2/wav"
)

type Control int

const (
	Play Control = iota
	TogglePlayback
	Stop
)

// Player holds the control and feed channels to communicate with the DJ
type Player struct {
	Control chan Control
	Feed    chan Track
	ctrl    *beep.Ctrl
}

func NewPlayer() *Player {
	player := &Player{
		Control: make(chan Control),
		Feed:    make(chan Track),
		ctrl:    nil,
	}

	go DJ(player)
	return player
}

func (p *Player) Play() {
	p.Control <- Play
}

func (p *Player) TogglePlayback() {
	p.Control <- TogglePlayback
}

func (p *Player) Stop() {
	p.Control <- Stop
}

func (p *Player) Load(track Track) {
	p.Feed <- track
}

func (p *Player) Close() {
	close(p.Control)
	close(p.Feed)
}

func DJ(player *Player) error {

	var streamer beep.StreamSeekCloser
	var format beep.Format

	for {
		select {
		case track := <-player.Feed:

			if streamer != nil {
				speaker.Clear()
			}

			f, err := os.Open(track.Path())
			if err != nil {
				return err
			}
			switch track.Format() {
			case FLAC:
				streamer, format, err = flac.Decode(f)
			case MP3:
				streamer, format, err = mp3.Decode(f)
			case WAV:
				streamer, format, err = wav.Decode(f)
			case OGG:
				streamer, format, err = vorbis.Decode(f)
			default:
				err = fmt.Errorf("unsupported format: %s", track.Path())
			}
			if err != nil {
				return err
			}
			defer streamer.Close()

			speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
			player.ctrl = &beep.Ctrl{Streamer: streamer}
		case cmd := <-player.Control:
			switch cmd {
			case Play:
				speaker.Play(player.ctrl)
				speaker.Lock()
				player.ctrl.Paused = false
				speaker.Unlock()
			case TogglePlayback:
				speaker.Lock()
				player.ctrl.Paused = !player.ctrl.Paused
				speaker.Unlock()
			case Stop:
				speaker.Clear()
				streamer.Seek(0)
			}
		}
	}
}
