//go:build !test

package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// GameState defines the high-level mode the game is currently in.
type GameState string

const (
	StateBoot     GameState = "BOOT"
	StatePlaying  GameState = "PLAYING"
	StateGameOver GameState = "GAMEOVER"
)

// Game holds all state, entities, and timing data for a single session.
type Game struct {
	State GameState

	SW, SH float32 // Screen dimensions cached per frame
	Audio  *AudioSystem

	BallTexture rl.Texture2D
	Camera      rl.Camera2D // Used to apply screen shake

	HighScore int
	NewPB     bool // Triggers fireworks if a new PB is set

	// Blob Physics state
	BallX, BallY             float32 // The leading head of the blob
	SmearX, SmearY           float32 // The trailing tail of the blob
	TailLength               float32 // Head to tail distance, already clamped
	TailAngle                float32 // Direction the tail dome points, in degrees
	IdleTargetX, IdleTargetY float32 // Where the blob rests when no word is active
	BlobScale                float32 // Size of the blob, grows as words are eaten

	Targets      []*Target
	Particles    []Particle
	ActiveTarget *Target // The specific word the player is currently typing

	SpawnTimer      float32
	Level           int
	Score           int
	Health          int
	SpeedMultiplier float32 // Increases difficulty over time
	WordsEaten      int

	// "Juice" timers for game feel
	ScreenShakeTimer    float32
	ScreenShakeDuration float32 // Total length of the current shake, used to fade it out
	HitStopTimer        float32 // Pauses the game loop briefly on impact for emphasis
	LevelUpTimer        float32 // Displays the level-up text overlay

	BootTimer       float32
	PlayedBootSound bool
}

// makeBallTexture procedurally generates a solid white circle texture in memory.
func makeBallTexture() rl.Texture2D {
	diameter := circleTextureRadius * 2
	img := rl.GenImageColor(int(diameter), int(diameter), rl.Blank)
	rl.ImageDrawCircle(img, int32(diameter/2), int32(diameter/2), int32(circleTextureRadius), rl.White)
	tex := rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	return tex
}

// NewGame initializes systems and returns a fresh game instance in the boot state.
func NewGame() *Game {
	g := &Game{
		Audio:       InitAudio(),
		BallTexture: makeBallTexture(),
	}
	g.Camera.Zoom = 1.0
	g.HighScore = 0
	g.BootTimer = 0.0
	g.PlayedBootSound = false

	g.Reset()
	g.State = StateBoot
	return g
}

// Close frees loaded assets from GPU memory.
func (g *Game) Close() {
	rl.UnloadTexture(g.BallTexture)
	g.Audio.Close()
}

// Reset clears out the board and restores default stats for a new run.
func (g *Game) Reset() {
	g.SW = float32(rl.GetScreenWidth())
	g.SH = float32(rl.GetScreenHeight())

	g.BallX, g.BallY = g.SW/2, g.SH/2
	g.SmearX, g.SmearY = g.SW/2, g.SH/2
	g.IdleTargetX, g.IdleTargetY = g.SW/2, g.SH/2

	g.BlobScale = 0.05
	g.Targets = nil
	g.Particles = nil
	g.SpawnTimer = 0.0
	g.ActiveTarget = nil

	g.Level = 1
	g.Score = 0
	g.Health = 5
	g.SpeedMultiplier = 1.0
	g.WordsEaten = 0

	g.ScreenShakeTimer = 0.0
	g.ScreenShakeDuration = 0.0
	g.HitStopTimer = 0.0
	g.LevelUpTimer = 0.0

	g.State = StatePlaying
	g.NewPB = false
}

// randomShakeOffset returns a random offset within ±strength pixels, used to
// jitter the camera. A hard limit of 0 means "no shake left" so the camera
// comes back to rest rather than sitting one pixel off-centre.
func randomShakeOffset(strength float32) float32 {
	magnitude := int32(strength)
	if magnitude <= 0 {
		return 0
	}
	return float32(rl.GetRandomValue(-magnitude, magnitude))
}

// startScreenShake kicks off a camera shake that lasts for the given number
// of seconds. The duration is remembered so the shake can fade out over its
// own lifetime instead of staying at full strength until it stops.
func (g *Game) startScreenShake(duration float32) {
	g.ScreenShakeTimer = duration
	g.ScreenShakeDuration = duration
}
