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

// PlaybackControl is an enum to control the players audio playback
type PlaybackControl int

const (
	Play PlaybackControl = iota
	TogglePlayback
	Stop
	Skip
)

// PlaybackUpdate is a struct to send playback updates to the UI as it changes
type PlaybackUpdate struct {
	CurrentTrack Track
	IsPlaying    bool
}

// StatusMessage is a struct to send status messages to the UI in response to actions such as enqueueing a track
type StatusMessage struct {
	Message string
}

type Player struct {
	PlaybackUpdate  chan PlaybackUpdate
	StatusMessage   chan StatusMessage
	playbackControl chan PlaybackControl
	trackFeed       chan Track
	trackQueue      chan Track
	readyToPlay     chan bool
	currentTrack    Track
	queue           []Track
	ctrl            *beep.Ctrl
	isPlaying       bool
}

func NewPlayer() *Player {
	player := &Player{
		PlaybackUpdate:  make(chan PlaybackUpdate),
		StatusMessage:   make(chan StatusMessage),
		playbackControl: make(chan PlaybackControl),
		trackFeed:       make(chan Track),
		trackQueue:      make(chan Track),
		readyToPlay:     make(chan bool),
		currentTrack:    Track{},
		queue:           []Track{},
		ctrl:            nil,
		isPlaying:       false,
	}

	go player.playLoop()
	return player
}

func (p *Player) Play() {
	p.playbackControl <- Play
}

func (p *Player) TogglePlayback() {
	p.playbackControl <- TogglePlayback
}

func (p *Player) Stop() {
	p.playbackControl <- Stop
}

func (p *Player) ForcePlay(track Track) {
	p.trackFeed <- track
}

func (p *Player) Enqueue(track Track) {
	p.trackQueue <- track
	p.StatusMessage <- StatusMessage{Message: fmt.Sprintf("Enqueued: %s", track.Title())}
}

func (p *Player) EnqueueAll(tracks []Track) {
	for _, track := range tracks {
		p.Enqueue(track)
	}
	p.StatusMessage <- StatusMessage{Message: fmt.Sprintf("Enqueued %d tracks", len(tracks))}
}

func (p *Player) ClearQueue() {
	p.queue = []Track{}
}

func (p *Player) Skip() {
	p.playbackControl <- Skip
}

func (p *Player) Close() {
	close(p.playbackControl)
	close(p.trackFeed)
}

func (p *Player) playLoop() error {

	var streamer beep.StreamSeekCloser
	var format beep.Format

	for {
		select {
		case track := <-p.trackQueue:
			// Add track to queue, if no track is playing, start the first track
			if !p.isPlaying {
				p.isPlaying = true
				go func() {
					p.trackFeed <- track
				}()
			} else {
				p.queue = append(p.queue, track)
			}

		case track := <-p.trackFeed:
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
			p.ctrl = &beep.Ctrl{Streamer: streamer}
			p.currentTrack = track

			go func() {
				p.Play()
			}()

		case cmd := <-p.playbackControl:
			switch cmd {
			case Play:
				// Start playback. Callback at the end of the track will start the next track in the queue
				speaker.Lock()
				p.ctrl.Paused = false
				p.isPlaying = true
				speaker.Unlock()
				speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
					// Track has finished playing, start the next track in the queue if there is one
					p.Skip()
				})))

				p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: p.currentTrack, IsPlaying: p.isPlaying}

			case TogglePlayback:
				// Swap play/pause state and update the UI
				speaker.Lock()
				p.ctrl.Paused = !p.ctrl.Paused
				speaker.Unlock()

				p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: p.currentTrack, IsPlaying: !p.ctrl.Paused}

			case Stop:
				// Stop all playback and clear the queue
				speaker.Clear()
				if streamer != nil {
					streamer.Seek(0)
				}
				p.isPlaying = false
				p.queue = []Track{}
				p.currentTrack = Track{}

				p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: Track{}, IsPlaying: p.isPlaying}

			case Skip:
				// Stop the current track and start the next track in the queue
				speaker.Clear()
				if streamer != nil {
					streamer.Seek(0)
				}

				p.isPlaying = false
				if len(p.queue) > 0 {
					track := p.queue[0]
					p.queue = p.queue[1:]
					go func() {
						p.trackFeed <- track
					}()
				} else {
					p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: Track{}, IsPlaying: p.isPlaying}
				}

			}
		}
	}
}
