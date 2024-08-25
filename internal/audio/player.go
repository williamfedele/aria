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

	// Player state
	queue         []Track
	queuePosition int
	audioSettings *audioSettings
	volumeLevel   float64
}

func NewPlayer() *Player {
	player := &Player{
		PlaybackUpdate: make(chan PlaybackUpdate),
		StatusMessage:  make(chan StatusMessage),

		queue:         []Track{},
		queuePosition: 0,
		audioSettings: nil,
		volumeLevel:   0,
	}

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
	// Start playback. Callback at the end of the track will start the next track in the queue

	if len(p.queue) == 0 {
		p.sendStatusMessage("Queue is empty")
		return
	}

	if p.audioSettings != nil &&
		p.audioSettings.streamer != nil {

		speaker.Clear()
	}

	streamer, format, err := decodeTrack(p.queue[p.queuePosition])
	if err != nil {
		p.sendStatusMessage(fmt.Sprintf("Error decoding track: %s", err))
		return
	}
	p.updateAudioSettings(streamer, format)

	speaker.Lock()
	p.audioSettings.ctrl.Paused = false
	speaker.Unlock()

	speaker.Play(beep.Seq(p.audioSettings.volume, beep.Callback(func() {
		p.Next()
	})))

	p.sendPlaybackUpdate()
}

func (p *Player) ForcePlay(track Track) {
	p.ClearQueue()
	p.queue = append(p.queue, track)
	p.Play()
}

func (p *Player) Next() {
	if p.queuePosition < len(p.queue)-1 {
		p.queuePosition++
		p.Play()
	} else {
		p.Stop()
	}
}
func (p *Player) Previous() {
	// If the current track is more than 5 seconds in, restart the track
	if p.audioSettings != nil &&
		p.audioSettings.streamer != nil &&
		p.queuePosition > 0 &&
		(float64(p.audioSettings.streamer.Position())/float64(p.audioSettings.streamer.Len()) < 0.10) {

		p.queuePosition--
	}
	p.Play()
}

func (p *Player) Enqueue(track Track) {
	p.queue = append(p.queue, track)
	if !p.isPlaying() {
		p.Play()
	} else {
		p.sendStatusMessage(fmt.Sprintf("Enqueued: %s", track.Title()))
	}
}

func (p *Player) EnqueueAll(tracks []Track) {
	p.queue = append(p.queue, tracks...)
	if !p.isPlaying() {
		p.Play()
	}
	p.sendStatusMessage(fmt.Sprintf("Enqueued %d tracks", len(tracks)))
}

func (p *Player) Stop() {
	// Stop all playback and clear the queue
	speaker.Clear()
	p.audioSettings = nil

	p.ClearQueue()
	p.sendPlaybackUpdate()
}

func (p *Player) TogglePlayback() {
	// Swap play/pause state and update the UI
	if p.audioSettings == nil {
		return
	}

	speaker.Lock()
	p.audioSettings.ctrl.Paused = !p.audioSettings.ctrl.Paused
	speaker.Unlock()

	p.sendPlaybackUpdate()
}

func (p *Player) ClearQueue() {
	p.queue = []Track{}
	p.queuePosition = 0
}

func (p *Player) VolumeUp() {
	p.volumeLevel += 0.5
	if p.audioSettings != nil {
		p.audioSettings.volume.Volume = p.volumeLevel
	}

	p.sendStatusMessage(fmt.Sprintf("Volume: %.1f", p.volumeLevel))
}

func (p *Player) VolumeDown() {
	p.volumeLevel -= 0.5
	if p.audioSettings != nil {
		p.audioSettings.volume.Volume = p.volumeLevel
	}

	p.sendStatusMessage(fmt.Sprintf("Volume: %.1f", p.volumeLevel))
}

func (p *Player) getCurrentPlaying() Track {
	if !p.isPlaying() {
		return Track{}
	}
	return p.queue[p.queuePosition]
}

func (p *Player) isPlaying() bool {
	return p.audioSettings != nil && !p.audioSettings.ctrl.Paused
}

func (p *Player) sendStatusMessage(message string) {
	go func() {
		p.StatusMessage <- StatusMessage{Message: message}
	}()
}

func (p *Player) sendPlaybackUpdate() {
	go func() {
		p.PlaybackUpdate <- PlaybackUpdate{CurrentTrack: p.getCurrentPlaying(), IsPlaying: p.isPlaying()}
	}()
}

func decodeTrack(track Track) (beep.StreamSeekCloser, beep.Format, error) {
	f, err := os.Open(track.Path())
	if err != nil {
		return nil, beep.Format{}, err
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
		return nil, beep.Format{}, err
	}

	return streamer, format, nil
}
