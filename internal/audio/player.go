package audio

import (
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/speaker"
)

type Control int

const (
	Play Control = iota
	Pause
	Stop
)

type Player struct {
	Control chan Control
	Feed    chan string
}

func NewPlayer() *Player {
	player := &Player{
		Control: make(chan Control),
		Feed:    make(chan string),
	}

	go Worker(player.Control, player.Feed)
	return player
}

func (p *Player) Play() {
	p.Control <- Play
}

func (p *Player) Pause() {
	p.Control <- Pause
}

func (p *Player) Stop() {
	p.Control <- Stop
}

func (p *Player) Load(path string) {
	p.Feed <- path
}

func (p *Player) Close() {
	close(p.Control)
	close(p.Feed)
}

func Worker(trackControl <-chan Control, trackFeed <-chan string) error {

	var streamer beep.StreamSeekCloser
	var format beep.Format

	for {
		select {
		case path := <-trackFeed:

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
		case cmd := <-trackControl:
			switch cmd {
			case Play:
				speaker.Play(streamer)
			case Pause:
				speaker.Lock()
				speaker.Unlock()
			case Stop:
				speaker.Clear()
				streamer.Seek(0)
			}
		}
	}
}
