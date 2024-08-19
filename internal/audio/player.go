package audio

import (
	"fmt"
	"os"
	"sync"
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
	Next
)

// Player holds the control and feed channels to communicate with the DJ
type Player struct {
	Control   chan Control
	Feed      chan Track
	Queue     chan Track
	Ready     chan bool
	ctrl      *beep.Ctrl
	isPlaying bool
	mu        sync.Mutex
}

func NewPlayer() *Player {
	player := &Player{
		Control: make(chan Control),
		Feed:    make(chan Track),
		Queue:   make(chan Track),
		Ready:   make(chan bool),
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

func (p *Player) Enqueue(track Track) {
	p.Queue <- track
}

func (p *Player) Close() {
	close(p.Control)
	close(p.Feed)
}

func DJ(player *Player) error {

	var streamer beep.StreamSeekCloser
	var format beep.Format
	var queue []Track
	done := make(chan struct{})

	for {
		select {
		case track := <-player.Queue:
			// Add track to queue, if no track is playing, start the first track
			queue = append(queue, track)
			player.mu.Lock()
			if !player.isPlaying && len(queue) > 0 {
				player.isPlaying = true
				next := queue[0]
				go func() {
					player.Feed <- next
				}()
				queue = queue[1:]
			}
			player.mu.Unlock()

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
			// TODO send update to UI to display next track
			go func() {
				player.Control <- Play
			}()

		case <-done:
			// Track has finished playing, start the next track in the queue
			if len(queue) > 0 {
				track := queue[0]
				queue = queue[1:]
				go func() {
					player.Feed <- track
				}()
			}
		case cmd := <-player.Control:
			switch cmd {
			case Play:
				speaker.Lock()
				player.ctrl.Paused = false
				speaker.Unlock()
				// TODO set isPlaying to false when the track finishes
				speaker.Play(beep.Seq(player.ctrl, beep.Callback(func() {
					done <- struct{}{}
				})))

			case TogglePlayback:
				speaker.Lock()
				player.ctrl.Paused = !player.ctrl.Paused
				speaker.Unlock()
			case Stop:
				// TODO stop should clear the queue
				speaker.Clear()
				streamer.Seek(0)
			}
		}
	}
}
