package model

import (
	"io"
	"io/fs"
	"lintech/rego/game/loader"
	"log"
	"path"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/harbdog/raycaster-go/geom"
)

const (
	audioContextSampleRate = 48000
)

var audioContext = audio.NewContext(audioContextSampleRate)

type RegoAudioPlayer struct {
	player    *audio.Player
	audioFile fs.File
}

func LoadAudioPlayer(fname string) *RegoAudioPlayer {
	if len(fname) == 0 {
		return nil
	}

	f := loader.LoadAudioFile(fname)
	var d io.Reader
	var err error
	switch ext := path.Ext(fname); ext {
	case ".wav":
		d, err = wav.DecodeWithSampleRate(audioContextSampleRate, f)
	case ".mp3":
		d, err = mp3.DecodeWithSampleRate(audioContextSampleRate, f)
	default:
		log.Fatal("Unsupported audo file ext ", ext)
	}
	if err != nil {
		log.Fatalf("Decode audio file fail: %e", err)
	}
	// Create an audio.Player that has one stream.
	audioPlayer, err := audioContext.NewPlayer(d)
	if err != nil {
		log.Fatalf("Create audio player fail: %e", err)
	}
	return &RegoAudioPlayer{player: audioPlayer, audioFile: f}
}

func (a *RegoAudioPlayer) Close() {
	a.player.Close()
	a.audioFile.Close()
}

func (a *RegoAudioPlayer) Play(audioPosition, playerPosition Position, renderAudioDistance float64) {
	if a.player.IsPlaying() {
		return
	}
	if err := a.player.Rewind(); err != nil {
		log.Printf("Warning: Audioplayer Rewind fail!")
	} else {
		distance := geom.Distance2(playerPosition.X, playerPosition.Y,
			audioPosition.X, audioPosition.Y)
		if distance < renderAudioDistance {
			volume := 1.0 - distance/renderAudioDistance
			log.Printf(" Audioplayer distance: %v/%v, volume: %v", distance, renderAudioDistance, volume)
			a.player.SetVolume(volume)
			a.player.Play()
		}
	}
}

func (a *RegoAudioPlayer) PlayWithVolume(volume float64, forcePlay bool) {
	if a.player.IsPlaying() && !forcePlay {
		return
	}
	if err := a.player.Rewind(); err != nil {
		log.Printf("Warning: Audioplayer Rewind fail!")
	} else {
		a.player.SetVolume(volume)
		a.player.Play()
	}
}
