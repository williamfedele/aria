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

// PlaybackControl is an enum to control the players audio playback
type PlaybackControl int

const (
	Play PlaybackControl = iota
	TogglePlayback
	Stop
	Next
)

// PlaybackUpdate is a struct to send updates to the UI
type PlaybackUpdate struct {
	CurrentTrack Track
	IsPlaying    bool
}

// Player holds the control and feed channels to communicate with the DJ
type Player struct {
	PlaybackUpdate  chan PlaybackUpdate
	playbackControl chan PlaybackControl
	trackFeed       chan Track
	trackQueue      chan Track
	readyToPlay     chan bool
	queue           []Track
	ctrl            *beep.Ctrl
	isPlaying       bool
	mu              sync.Mutex
}

func NewPlayer() *Player {
	player := &Player{
		playbackControl: make(chan PlaybackControl),
		trackFeed:       make(chan Track),
		trackQueue:      make(chan Track),
		readyToPlay:     make(chan bool),
		PlaybackUpdate:  make(chan PlaybackUpdate),
		queue:           []Track{},
		ctrl:            nil,
	}

	go DJ(player)
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
	p.queue = []Track{track}
	p.trackFeed <- track
}

func (p *Player) Enqueue(track Track) {
	p.trackQueue <- track
}

func (p *Player) Close() {
	close(p.playbackControl)
	close(p.trackFeed)
}

func DJ(player *Player) error {

	var streamer beep.StreamSeekCloser
	var format beep.Format
	//done := make(chan struct{})
	currentTrack := Track{}

	for {
		select {
		case track := <-player.trackQueue:
			// Add track to queue, if no track is playing, start the first track
			if !player.isPlaying {
				player.isPlaying = true
				go func() {
					player.trackFeed <- track
				}()
			} else {
				player.queue = append(player.queue, track)
			}

		case track := <-player.trackFeed:
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
			currentTrack = track

			go func() {
				player.playbackControl <- Play
			}()

		// case <-done:
		// 	// Track has finished playing, start the next track in the queue
		// 	if len(queue) > 0 {
		// 		track := queue[0]
		// 		queue = queue[1:]
		// 		go func() {
		// 			player.trackFeed <- track
		// 		}()
		// 	}
		case cmd := <-player.playbackControl:
			switch cmd {
			case Play:
				// Start playback. Callback at the end of the track will start the next track in the queue
				speaker.Lock()
				player.ctrl.Paused = false
				speaker.Unlock()
				speaker.Play(beep.Seq(player.ctrl, beep.Callback(func() {
					// Track has finished playing, start the next track in the queue if there is one
					if len(player.queue) > 0 {
						track := player.queue[0]
						player.queue = player.queue[1:]
						go func() {
							player.trackFeed <- track
						}()
					} else {
						player.isPlaying = false
						player.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: Track{}, IsPlaying: false}
					}
				})))
				go func() {
					//fmt.Println("Send update: ", currentTrack.Title())
					player.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: currentTrack, IsPlaying: true}

					//fmt.Println("Sent update: ", currentTrack.Title())
				}()

			case TogglePlayback:
				// Swap play/pause state and update the UI
				speaker.Lock()
				player.ctrl.Paused = !player.ctrl.Paused
				player.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: currentTrack, IsPlaying: !player.ctrl.Paused}
				speaker.Unlock()
			case Stop:
				// Stop all playback and clear the queue
				speaker.Clear()
				streamer.Seek(0)
				player.isPlaying = false
				player.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: Track{}, IsPlaying: false}
				player.queue = []Track{}
			}
		}
	}
}
