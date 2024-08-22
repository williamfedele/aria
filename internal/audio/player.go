package audio

import (
	"fmt"
	"os"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
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

type audioSettings struct {
	streamer beep.StreamSeekCloser
	ctrl     *beep.Ctrl
	volume   *effects.Volume
}

type Player struct {
	// Channels for sending updates to the UI
	PlaybackUpdate chan PlaybackUpdate
	StatusMessage  chan StatusMessage

	// Channels for controlling the player
	playbackControl chan PlaybackControl
	trackFeed       chan Track
	trackQueue      chan Track

	// Player state
	currentTrack  Track
	queue         []Track
	audioSettings *audioSettings
	volumeLevel   float64
	isPlaying     bool
}

func NewPlayer() *Player {
	player := &Player{
		PlaybackUpdate: make(chan PlaybackUpdate),
		StatusMessage:  make(chan StatusMessage),

		playbackControl: make(chan PlaybackControl),
		trackFeed:       make(chan Track),
		trackQueue:      make(chan Track),

		currentTrack:  Track{},
		queue:         []Track{},
		audioSettings: nil,
		volumeLevel:   0,
		isPlaying:     false,
	}

	go player.playLoop()
	return player
}

func (p *Player) updateAudioSettings(streamer beep.StreamSeekCloser, format beep.Format) {
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	ctrl := &beep.Ctrl{Streamer: streamer}
	volume := &effects.Volume{Streamer: ctrl, Base: 2, Volume: p.volumeLevel}

	p.audioSettings = &audioSettings{
		streamer: streamer,
		ctrl:     ctrl,
		volume:   volume,
	}
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

func (p *Player) VolumeUp() {
	p.volumeLevel += 0.5
	if p.audioSettings != nil {
		p.audioSettings.volume.Volume = p.volumeLevel
	}

	p.StatusMessage <- StatusMessage{Message: fmt.Sprintf("Volume: %.1f", p.volumeLevel)}
}

func (p *Player) VolumeDown() {
	p.volumeLevel -= 0.5
	if p.audioSettings != nil {
		p.audioSettings.volume.Volume = p.volumeLevel
	}

	p.StatusMessage <- StatusMessage{Message: fmt.Sprintf("Volume: %.1f", p.volumeLevel)}
}

func (p *Player) playLoop() error {

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
			if p.audioSettings != nil && p.audioSettings.streamer != nil {
				speaker.Clear()
			}

			f, err := os.Open(track.Path())
			if err != nil {
				return err
			}

			var streamer beep.StreamSeekCloser
			var format beep.Format

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

			p.updateAudioSettings(streamer, format)
			p.currentTrack = track

			go func() {
				p.Play()
			}()

		case cmd := <-p.playbackControl:
			switch cmd {
			case Play:
				// Start playback. Callback at the end of the track will start the next track in the queue
				speaker.Lock()
				p.audioSettings.ctrl.Paused = false
				p.isPlaying = true
				speaker.Unlock()
				speaker.Play(beep.Seq(p.audioSettings.volume, beep.Callback(func() {
					// Track has finished playing, start the next track in the queue if there is one
					p.Skip()
				})))

				p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: p.currentTrack, IsPlaying: p.isPlaying}

			case TogglePlayback:
				// Swap play/pause state and update the UI
				speaker.Lock()
				p.audioSettings.ctrl.Paused = !p.audioSettings.ctrl.Paused
				speaker.Unlock()

				p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: p.currentTrack, IsPlaying: !p.audioSettings.ctrl.Paused}

			case Stop:
				// Stop all playback and clear the queue
				speaker.Clear()
				if p.audioSettings != nil && p.audioSettings.streamer != nil {
					p.audioSettings.streamer.Seek(0)
				}
				p.isPlaying = false
				p.queue = []Track{}
				p.currentTrack = Track{}

				p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: Track{}, IsPlaying: p.isPlaying}

			case Skip:
				// Stop the current track and start the next track in the queue
				speaker.Clear()
				if p.audioSettings != nil && p.audioSettings.streamer != nil {
					p.audioSettings.streamer.Seek(0)
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
