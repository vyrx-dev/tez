//go:build !test

package main

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	// circleTextureRadius is the base resolution for the procedurally generated blob texture.
	circleTextureRadius = 256
	// baseBallChaseSpeed defines how aggressively the main blob tracks the target.
	baseBallChaseSpeed float32 = 15.0
	// baseSmearChaseSpeed is slightly slower than the ball, creating the stretchy tail effect.
	baseSmearChaseSpeed float32 = baseBallChaseSpeed * 0.45
	// maxScreenShakeOffset is how far the camera can be thrown, in pixels,
	// at the very start of a shake.
	maxScreenShakeOffset = 10
)

// Gameplay tuning. Mostly magic numbers that felt right, grouped here so they
// can be found without reading the update logic.
const (
	wordsPerLevel = 10 // Words eaten before the level goes up

	levelSpeedBonus           float32 = 0.3  // Added to SpeedMultiplier per level
	levelHitStopDuration      float32 = 0.05 // Seconds the world freezes on level-up
	levelUpBannerDuration     float32 = 2.0  // Seconds the "LEVEL N" text stays up
	wordFallSpeed             float32 = 40   // Pixels per second a word falls at level 1
	particleGravity           float32 = 800  // Pixels per second squared, pulls sparks down
	eatenWordParticleCount            = 30   // Sparks when a word is eaten
	personalBestFireworkCount         = 200  // Sparks for the new-personal-best celebration
	targetPopSpeed            float32 = 15.0 // How fast a word snaps back after being hit
	wordPopScale              float32 = 1.5  // How much a word inflates when hit
	maxBlobScale              float32 = 1.5  // The blob stops growing at this scale
	bootIntroDuration         float32 = 2.0  // Seconds the "Tez" title card shows for
)

var wordList = []string{
	// Easy
	"cat", "dog", "run", "box", "car", "bus", "sun", "sky", "pen", "hat",
	"jump", "ball", "tree", "fish", "bird", "moon", "star", "play", "game", "fast",
	"cool", "blue", "red", "door", "desk", "book", "ship", "shoe", "frog", "bear",
	// Medium
	"yellow", "purple", "spring", "summer", "winter", "autumn", "school", "pencil", "planet", "rocket",
	"dragon", "knight", "castle", "magic", "sword", "shield", "wizard", "potion",
	// Hard
	"adventure", "beautiful", "challenge", "dangerous", "elephant", "fantastic", "mountain", "telescope", "universe", "waterfall",
}

// Target represents a falling word that the player must type.
type Target struct {
	X, Y       float32 // Screen coordinates
	Word       string  // The complete word
	TypedIndex int     // How many characters have been correctly typed so far
	Active     bool    // If false, the word is destroyed and removed
	Scale      float32 // Used for the bouncy pop animation when typing a letter
}

// Particle represents a physics-driven spark in an explosion.
type Particle struct {
	X, Y    float32
	VX, VY  float32 // Velocity vectors
	Life    float32 // Current remaining lifetime in seconds
	MaxLife float32 // Starting lifetime, used to calculate alpha fading
	Color   rl.Color
	Size    float32
}
