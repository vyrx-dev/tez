//go:build !test

package main

import (
	_ "embed"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The sound effects are embedded rather than written out at startup.
//
// They used to be synthesised into ./audio when the game started, which meant
// the game had to be launched from a writable directory. Running it from
// somewhere read-only (or from /usr/bin after installing it) silently produced
// no sound at all, because the file creation failed and nothing reported it.
// Embedding the files means there is nothing to write and nothing to lose.
//
// They were generated once by a small waveform synthesiser; the results are
// checked into assets/audio, so add new sounds there instead of at runtime.
var (
	//go:embed assets/audio/type.wav
	typeSound []byte

	//go:embed assets/audio/eat.wav
	eatSound []byte

	//go:embed assets/audio/hurt.wav
	hurtSound []byte

	//go:embed assets/audio/levelup.wav
	levelUpSound []byte

	//go:embed assets/audio/pb.wav
	personalBestSound []byte

	//go:embed assets/audio/boot.wav
	bootSound []byte

	//go:embed assets/audio/gameover.wav
	gameOverSound []byte
)

// soundFiles maps the names the game asks for to their embedded bytes, so
// adding a sound is one entry here plus one file in assets/audio.
var soundFiles = map[string][]byte{
	"type":     typeSound,
	"eat":      eatSound,
	"hurt":     hurtSound,
	"levelup":  levelUpSound,
	"pb":       personalBestSound,
	"boot":     bootSound,
	"gameover": gameOverSound,
}

// AudioSystem holds the sounds the game plays, loaded into raylib once at
// startup.
type AudioSystem struct {
	Sounds map[string]rl.Sound
}

// InitAudio loads the embedded sounds into raylib.
func InitAudio() *AudioSystem {
	sys := &AudioSystem{
		Sounds: make(map[string]rl.Sound),
	}

	for name, data := range soundFiles {
		sys.Sounds[name] = loadSound(data)
	}

	return sys
}

// loadSound turns embedded WAV bytes into a raylib sound.
func loadSound(data []byte) rl.Sound {
	wave := rl.LoadWaveFromMemory(".wav", data, int32(len(data)))

	sound := rl.LoadSoundFromWave(wave)
	if !rl.IsSoundValid(sound) {
		// A zeroed sound means the bytes were not a usable wav. Nothing to do
		// but keep going, but say so rather than fail quietly.
		fmt.Println("audio: could not load an embedded sound")
	}

	// The wave's sample data is copied into the sound, so the wave can be
	// unloaded straight away.
	rl.UnloadWave(wave)

	return sound
}

// Close frees the audio assets from memory.
func (a *AudioSystem) Close() {
	for _, sound := range a.Sounds {
		rl.UnloadSound(sound)
	}
}

// Play triggers a sound by its string identifier.
func (a *AudioSystem) Play(name string) {
	if sound, ok := a.Sounds[name]; ok {
		rl.PlaySound(sound)
	}
}
