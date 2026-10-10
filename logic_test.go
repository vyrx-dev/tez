package main

import (
	"math"
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name           string
		x1, y1, x2, y2 float32
		expected       float32
	}{
		{"same point", 0, 0, 0, 0, 0},
		{"3-4-5 triangle", 0, 0, 3, 4, 5},
		{"horizontal line", 0, 5, 10, 5, 10},
		{"vertical line", 7, 0, 7, 24, 24},
		{"negative coordinates", -3, -4, 0, 0, 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := distance(tc.x1, tc.y1, tc.x2, tc.y2)
			if math.Abs(float64(got-tc.expected)) > 1e-5 {
				t.Errorf("distance(%v,%v, %v,%v) = %v, want %v",
					tc.x1, tc.y1, tc.x2, tc.y2, got, tc.expected)
			}
		})
	}
}

func TestDistanceIsSymmetric(t *testing.T) {
	pairs := [][4]float32{
		{1, 2, 5, 6},
		{-10, 3, 4, -7},
		{100, 200, 300, 400},
	}
	for _, pair := range pairs {
		fromAtoB := distance(pair[0], pair[1], pair[2], pair[3])
		fromBtoA := distance(pair[2], pair[3], pair[0], pair[1])
		if fromAtoB != fromBtoA {
			t.Errorf("distance not symmetric for %v: got %v and %v", pair, fromAtoB, fromBtoA)
		}
	}
}

func TestDistanceIsNeverNegative(t *testing.T) {
	pairs := [][4]float32{
		{0, 0, 0, 0},
		{-5, -5, 5, 5},
		{1000, -1000, -1000, 1000},
	}
	for _, pair := range pairs {
		if result := distance(pair[0], pair[1], pair[2], pair[3]); result < 0 {
			t.Errorf("got negative distance %v for %v", result, pair)
		}
	}
}

func TestApproach(t *testing.T) {
	tests := []struct {
		name      string
		current   float32
		target    float32
		deltaTime float32
		speed     float32
		expected  float32
	}{
		// already there, should stay put
		{"already at target", 5, 5, 0.016, 15, 5},

		// factor = 1-exp(-1.0*0.1) = 0.0951626 -> 0.951626
		{"simple step", 0, 10, 0.1, 1.0, 0.951626},

		// factor = 1-exp(-15*0.016) = 0.213372 -> 100 - 21.3372 = 78.6628
		{"approaching from above", 100, 0, 0.016, 15, 78.6628},

		// factor = 1-exp(-15*0.016) = 0.213372 -> -10.6686
		{"negative target", 0, -50, 0.016, 15, -10.6686},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := approach(tc.current, tc.target, tc.deltaTime, tc.speed)
			if math.Abs(float64(got-tc.expected)) > 1e-4 {
				t.Errorf("approach(%v, %v, %v, %v) = %v, want %v",
					tc.current, tc.target, tc.deltaTime, tc.speed, got, tc.expected)
			}
		})
	}
}

// The whole point of the exponential form: the same elapsed second closes the
// same share of the gap at any frame rate. Chase for 1s at 30, 60 and 144Hz -
// all three must land in nearly the same place.
func TestApproachIsFrameRateIndependent(t *testing.T) {
	target := float32(400)

	var distancesCovered []float32
	for _, frameRate := range []int{30, 60, 144} {
		deltaTime := float32(1.0) / float32(frameRate)
		current := float32(0)
		for frame := 0; frame < frameRate; frame++ {
			current = approach(current, target, deltaTime, 15)
		}
		distancesCovered = append(distancesCovered, current)
	}

	nearest, furthest := distancesCovered[0], distancesCovered[0]
	for _, covered := range distancesCovered {
		if covered < nearest {
			nearest = covered
		}
		if covered > furthest {
			furthest = covered
		}
	}

	if furthest-nearest > 1 {
		t.Errorf("1s of chasing landed in very different places at different frame rates: %v", distancesCovered)
	}
	if math.Abs(float64(furthest-target)) > 1 {
		t.Errorf("after 1s of chasing only reached %v of %v", furthest, target)
	}
}

// simulate 300 frames and make sure the ball never flies past the target
func TestApproachNeverOvershoots(t *testing.T) {
	currentPosition := float32(0)
	targetPosition := float32(400)

	for frame := 0; frame < 300; frame++ {
		next := approach(currentPosition, targetPosition, 1.0/60.0, 15.0)
		if next > targetPosition {
			t.Fatalf("frame %d: overshot target %.4f (got %.4f)", frame, targetPosition, next)
		}
		currentPosition = next
	}

	// should be basically at the target after ~5 simulated seconds
	if math.Abs(float64(currentPosition-targetPosition)) > 0.01 {
		t.Errorf("didn't converge: stuck at %v, target was %v", currentPosition, targetPosition)
	}
}

func TestApproachFactor(t *testing.T) {
	// No time passes, no gap closes.
	if factor := approachFactor(0, 15); factor != 0 {
		t.Errorf("approachFactor(0, 15) = %v, want 0", factor)
	}

	// A slower chase covers a smaller share per frame, but both stay under 1.
	slow := approachFactor(1.0/60.0, 5)
	fast := approachFactor(1.0/60.0, 15)
	if !(0 < slow && slow < fast && fast < 1) {
		t.Errorf("expected 0 < %v < %v < 1", slow, fast)
	}
}

func TestTailAngle(t *testing.T) {
	tests := []struct {
		name           string
		deltaX, deltaY float32
		expected       float32
	}{
		// atan2 works in radians; +90 because raylib's 0 points up
		{"right", 1, 0, 90},
		{"down", 0, 1, 180},
		{"up", 0, -1, 0},
		{"left", -1, 0, 270},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tailAngle(tc.deltaX, tc.deltaY)
			if math.Abs(float64(got-tc.expected)) > 1e-3 {
				t.Errorf("tailAngle(%v, %v) = %v, want %v", tc.deltaX, tc.deltaY, got, tc.expected)
			}
		})
	}
}

// The dome must swing the short way, including across the 0/360 seam.
func TestDampAngleTakesShortestPath(t *testing.T) {
	// 350 -> 10 is +20 the short way, not -340 the long way.
	if got := dampAngle(350, 10, 1.0/60.0, 18); !(got > 350 && got < 370) {
		t.Errorf("dampAngle(350, 10) = %v, expected it to cross 360 upward", got)
	}

	// ...and the same going down through zero.
	if got := dampAngle(-350, -10, 1.0/60.0, 18); !(got < -350 && got > -370) {
		t.Errorf("dampAngle(-350, -10) = %v, expected it to cross -360 downward", got)
	}
}

// A long tail must be capped, and a stretched tail must also get thinner.
func TestTailDimensions(t *testing.T) {
	t.Run("idle tail has no length", func(t *testing.T) {
		if length, _ := tailDimensions(0, 100); length != 0 {
			t.Errorf("length = %v, want 0", length)
		}
	})

	t.Run("one blob-width long matches the blob", func(t *testing.T) {
		length, width := tailDimensions(100, 100)
		if length != 100 || width != 100 {
			t.Errorf("got (%v, %v), want (100, 100)", length, width)
		}
	})

	t.Run("stretching keeps the ball's width", func(t *testing.T) {
		// This is the whole look of the effect: the dome stays as fat as the
		// ball no matter how far it is stretched, so it reads as a blob being
		// dragged along rather than a rubber band being pulled.
		shortLength, shortWidth := tailDimensions(100, 100)
		longLength, longWidth := tailDimensions(400, 100)

		if shortWidth != 100 || longWidth != 100 {
			t.Errorf("width went from %v to %v, want it pinned at 100", shortWidth, longWidth)
		}
		if math.Abs(float64(longLength-4*shortLength)) > 1e-4 {
			t.Errorf("length grew to %v, want %v", longLength, 4*shortLength)
		}
	})

	t.Run("tail is capped at maxTailStretch widths", func(t *testing.T) {
		// The idle target teleports, so the raw distance can be huge.
		longestPossible := maxTailStretch * 100
		if length, _ := tailDimensions(10000, 100); math.Abs(float64(length-longestPossible)) > 1e-4 {
			t.Errorf("length = %v, want it capped at %v", length, longestPossible)
		}
	})

	t.Run("zero-sized blob draws nothing", func(t *testing.T) {
		if length, width := tailDimensions(50, 0); length != 0 || width != 0 {
			t.Errorf("got (%v, %v), want (0, 0)", length, width)
		}
	})
}

func BenchmarkDistance(b *testing.B) {
	for b.Loop() {
		distance(10, 20, 30, 40)
	}
}

func BenchmarkApproach(b *testing.B) {
	pos := float32(0)
	for b.Loop() {
		pos = approach(pos, 400, 0.016, 15)
	}
}
