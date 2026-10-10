//go:build !test

package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// tickWordPop snaps a word back to its normal size after the little bounce it
// does when you hit a letter. Same easing as the blob's chase so it behaves
// identically at any frame rate.
func tickWordPop(target *Target, deltaTime float32) {
	target.Scale = approach(target.Scale, 1.0, deltaTime, targetPopSpeed)
}

// spawnEatBurst pops a handful of sparks out of a word you just finished,
// tinted with the current level's colour.
func (g *Game) spawnEatBurst(word *Target) {
	for i := 0; i < eatenWordParticleCount; i++ {
		angle := float64(rl.GetRandomValue(0, 360)) * math.Pi / 180.0
		speed := float32(rl.GetRandomValue(100, 500))
		g.Particles = append(g.Particles, Particle{
			X: word.X, Y: word.Y,
			VX:      float32(math.Cos(angle)) * speed,
			VY:      float32(math.Sin(angle)) * speed,
			Life:    1.0,
			MaxLife: 1.0,
			Color:   rl.ColorFromHSV(g.levelHue(), 0.8, 1.0),
			Size:    float32(rl.GetRandomValue(3, 10)),
		})
	}
}

// spawnPersonalBestBurst throws a big one-off firework display for beating your
// own high score.
func (g *Game) spawnPersonalBestBurst() {
	for i := 0; i < personalBestFireworkCount; i++ {
		angle := float64(rl.GetRandomValue(0, 360)) * math.Pi / 180.0
		speed := float32(rl.GetRandomValue(100, 1000))
		g.Particles = append(g.Particles, Particle{
			X: g.SW / 2, Y: g.SH / 2,
			VX: float32(math.Cos(angle)) * speed,
			VY: float32(math.Sin(angle)) * speed,
			// Each spark lives a slightly different amount of time so the
			// whole display fades out instead of blinking off all at once.
			Life:    float32(rl.GetRandomValue(10, 30)) / 10.0,
			MaxLife: 3.0,
			Color:   rl.ColorFromHSV(float32(rl.GetRandomValue(0, 360)), 1.0, 1.0),
			Size:    float32(rl.GetRandomValue(5, 15)),
		})
	}
}

// levelHue returns the current level's colour, wrapping around the wheel so a
// long run does not run out of hue.
func (g *Game) levelHue() float32 {
	return float32((g.Level * 40) % 360)
}

// Update runs exactly once per frame. It does no work itself beyond keeping
// the cached screen size current and dispatching to whichever state the game
// is in; the real per-frame logic lives in this file, one method per state.
func (g *Game) Update(dt float32) {
	g.SW = float32(rl.GetScreenWidth())
	g.SH = float32(rl.GetScreenHeight())

	switch g.State {
	case StateBoot:
		g.updateBoot(dt)
	case StatePlaying:
		g.updatePlaying(dt)
	case StateGameOver:
		g.updateGameOver(dt)
	}
}

// updateBoot waits out the intro card and then starts a run.
func (g *Game) updateBoot(dt float32) {
	if !g.PlayedBootSound {
		g.Audio.Play("boot")
		g.PlayedBootSound = true
	}

	g.BootTimer += dt
	if g.BootTimer > bootIntroDuration {
		g.Reset()
	}
}

// updateGameOver shows the explosion and waits for the player to restart.
func (g *Game) updateGameOver(dt float32) {
	g.updateParticles(dt)

	if rl.IsKeyPressed(rl.KeyEnter) {
		g.Reset()
	}
}

// updatePlaying is the game itself: timers, particles, spawning, typing, and
// then the blob. Ordered cheapest-first-ish so the interesting stuff is at the
// bottom, where you can read the blob chase without scrolling past everything.
func (g *Game) updatePlaying(dt float32) {
	if g.Health <= 0 {
		g.beginGameOver()
		return
	}

	// Hit-stop is a game feel trick: freeze the world for a few frames on a
	// heavy impact so it lands. It is deliberately kept rare (see updateWord
	// completion) - a pause on every word fired several times a second would
	// turn the stretchy tail into a stutter.
	if g.HitStopTimer > 0 {
		g.HitStopTimer -= dt
		return // Skip the rest of the update
	}

	if g.LevelUpTimer > 0 {
		g.LevelUpTimer -= dt
	}

	g.updateCameraShake(dt)
	g.updateParticles(dt)

	for _, target := range g.Targets {
		tickWordPop(target, dt)
	}

	g.spawnWordIfDue(dt)
	g.handleKeyPress()
	g.updateWords(dt)

	g.moveBlobTowardsWord(dt)
}

// beginGameOver switches to the game over screen, remembering whether this run
// beat the old high score so the renderer can celebrate it.
func (g *Game) beginGameOver() {
	g.State = StateGameOver

	// The `HighScore > 0` guard means the very first run of a fresh install
	// can never set a personal best - there is nothing to beat yet.
	if g.Score > g.HighScore && g.Score > 0 && g.HighScore > 0 {
		g.NewPB = true
		g.Audio.Play("pb")
		g.spawnPersonalBestBurst()
		g.HighScore = g.Score
		return
	}

	g.Audio.Play("gameover")
	if g.Score > g.HighScore {
		g.HighScore = g.Score
	}
}

// updateCameraShake throws the camera around while a shake is active, easing
// off towards the end rather than jittering at full strength until it stops.
func (g *Game) updateCameraShake(dt float32) {
	if g.ScreenShakeTimer <= 0 {
		g.Camera.Offset.X = 0
		g.Camera.Offset.Y = 0
		return
	}

	g.ScreenShakeTimer -= dt

	// Scale the strength by how much of the shake is left. A shake that stays
	// strong the whole time reads as noise; one that decays reads as an impact
	// the camera is settling down from.
	remaining := g.ScreenShakeTimer
	if remaining > g.ScreenShakeDuration {
		remaining = g.ScreenShakeDuration
	}
	strength := float32(maxScreenShakeOffset) * remaining / g.ScreenShakeDuration

	g.Camera.Offset.X = randomShakeOffset(strength)
	g.Camera.Offset.Y = randomShakeOffset(strength)
}

// updateParticles ages the explosion sparks and applies gravity to them.
func (g *Game) updateParticles(dt float32) {
	// Walked backwards so removing a dead particle cannot skip the next one.
	for i := len(g.Particles) - 1; i >= 0; i-- {
		particle := &g.Particles[i]

		particle.Life -= dt
		if particle.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
			continue
		}

		particle.X += particle.VX * dt
		particle.Y += particle.VY * dt
		particle.VY += particleGravity * dt
	}
}

// spawnWordIfDue drops a new word from the top of the screen on a timer that
// shortens as the level goes up.
func (g *Game) spawnWordIfDue(dt float32) {
	g.SpawnTimer -= dt
	if g.SpawnTimer > 0 {
		return
	}

	g.SpawnTimer = 2.0 / g.SpeedMultiplier

	// Kept away from the edges so a word never spawns half off-screen.
	leftMargin, rightMargin := float32(60), g.SW-60
	spawnX := float32(rl.GetRandomValue(int32(leftMargin), int32(rightMargin)))

	g.Targets = append(g.Targets, &Target{
		X:      spawnX,
		Y:      -50,
		Word:   wordList[rl.GetRandomValue(0, int32(len(wordList)-1))],
		Active: true,
		Scale:  1.0,
	})
}

// updateWords moves the falling words down and takes a life off when one
// reaches the bottom.
func (g *Game) updateWords(dt float32) {
	for _, target := range g.Targets {
		if !target.Active {
			continue
		}

		target.Y += wordFallSpeed * g.SpeedMultiplier * dt
		if target.Y < g.SH+30 {
			continue
		}

		// The word got past the player.
		target.Active = false
		g.Health--
		g.startScreenShake(0.4) // Losing a life is the heaviest hit
		g.Audio.Play("hurt")

		if g.ActiveTarget == target {
			g.ActiveTarget = nil
		}
	}
}

// handleKeyPress feeds the last key pressed into the typing logic. Only the
// letter keys matter, since the game asks for letters.
func (g *Game) handleKeyPress() {
	key := rl.GetKeyPressed()
	if key < rl.KeyA || key > rl.KeyZ {
		return
	}

	// raylib reports key codes in uppercase ASCII, and the words are stored
	// lowercase, so the key code is nudged down into the lowercase range.
	letter := byte(key + 32)

	if g.ActiveTarget != nil && g.ActiveTarget.Active {
		g.typeNextLetter(letter)
		return
	}

	g.beginWordStartingWith(letter)
}

// typeNextLetter checks the pressed letter against the word being typed.
func (g *Game) typeNextLetter(letter byte) {
	active := g.ActiveTarget
	if active.Word[active.TypedIndex] != letter {
		return // Wrong key, the word keeps its progress and waits
	}

	active.TypedIndex++
	active.Scale = wordPopScale
	g.Audio.Play("type")

	if active.TypedIndex < len(active.Word) {
		return
	}

	g.eatWord(active)
}

// beginWordStartingWith picks a word to work on out of the ones still on
// screen. The lowest one wins, because it is the one closest to hurting you.
func (g *Game) beginWordStartingWith(letter byte) {
	var chosen *Target
	var lowestOnScreen float32 = -1000

	for _, target := range g.Targets {
		if !target.Active || target.Word[0] != letter {
			continue
		}
		if target.Y > lowestOnScreen {
			lowestOnScreen = target.Y
			chosen = target
		}
	}

	if chosen == nil {
		return
	}

	g.ActiveTarget = chosen
	g.ActiveTarget.TypedIndex = 1
	g.ActiveTarget.Scale = wordPopScale
	g.Audio.Play("type")

	// A one-letter word is already finished after this press.
	if g.ActiveTarget.TypedIndex == len(g.ActiveTarget.Word) {
		g.eatWord(g.ActiveTarget)
	}
}

// eatWord removes a finished word and rewards the player.
func (g *Game) eatWord(word *Target) {
	// The blob now has nothing to chase, so it drifts to where the word was
	// instead of snapping back to the middle of the screen.
	g.IdleTargetX = word.X
	g.IdleTargetY = word.Y

	g.spawnEatBurst(word)

	word.Active = false
	if g.ActiveTarget == word {
		g.ActiveTarget = nil
	}

	g.Score += 10
	g.WordsEaten++
	g.BlobScale += 0.05
	g.Audio.Play("eat")

	// Eating a word shakes the camera but never freezes the world: that is
	// saved for level-ups, which happen rarely enough to feel special.
	g.startScreenShake(0.1)

	if g.WordsEaten%wordsPerLevel != 0 {
		if g.BlobScale > maxBlobScale {
			g.BlobScale = maxBlobScale
		}
		return
	}

	// Level up: faster words, and the blob goes back to its starting size so
	// it cannot grow to fill the screen over a long run.
	g.Level++
	g.BlobScale = 0.05
	g.SpeedMultiplier += levelSpeedBonus
	g.LevelUpTimer = levelUpBannerDuration
	g.Audio.Play("levelup")

	// The one heavy moment: freeze for a few frames so the level-up lands.
	g.HitStopTimer = levelHitStopDuration
	g.startScreenShake(0.1)
}

// moveBlobTowardsWord is the blob physics.
//
// Two points are chased each frame with exponential easing: the head runs
// after whichever word is being typed, and the tail runs after the head. The
// tail is the slower of the two, so whenever the head moves it leaves the tail
// behind, and that gap is what gets drawn as the stretched dome.
func (g *Game) moveBlobTowardsWord(dt float32) {
	targetX, targetY := g.IdleTargetX, g.IdleTargetY
	if g.ActiveTarget != nil && g.ActiveTarget.Active {
		targetX, targetY = g.ActiveTarget.X, g.ActiveTarget.Y
	}

	headSpeed := baseBallChaseSpeed * g.SpeedMultiplier
	tailSpeed := baseSmearChaseSpeed * g.SpeedMultiplier

	g.BallX = approach(g.BallX, targetX, dt, headSpeed)
	g.BallY = approach(g.BallY, targetY, dt, headSpeed)

	g.SmearX = approach(g.SmearX, g.BallX, dt, tailSpeed)
	g.SmearY = approach(g.SmearY, g.BallY, dt, tailSpeed)

	// Record how far the tail lags so the renderer knows how long to draw the
	// dome, and aim it along that direction. The angle is eased rather than
	// taken straight from Atan2, so the dome swings into its new heading
	// instead of snapping to it.
	g.TailLength = distance(g.BallX, g.BallY, g.SmearX, g.SmearY)
	if g.TailLength <= tailDeadZone {
		return
	}

	aim := tailAngle(g.SmearX-g.BallX, g.SmearY-g.BallY)
	g.TailAngle = dampAngle(g.TailAngle, aim, dt, tailAngleSpeed)
}
