package main

import "math"

// degreesPerRadian converts radians to degrees. math.Atan2 works in radians
// and raylib rotations are in degrees.
const degreesPerRadian = float32(180.0 / math.Pi)

// Tuning for the stretchy tail that trails behind the blob. These live here,
// next to the tail maths, because they only affect the blob's shape.
const (
	// The dome is capped at maxTailStretch blob-widths long. The cap matters
	// because the idle target teleports to whichever word was just eaten, so
	// uncapped the smear can stretch across the whole screen.
	maxTailStretch float32 = 4.0

	// Hides the dome while head and tail sit closer than this, because at
	// those lengths the direction between the two is mostly rounding noise
	// and a live angle would flicker.
	tailDeadZone float32 = 4.0

	// How much of the remaining angle the dome covers per second. High snaps
	// straight to the new heading, low lags behind it and looks dragged.
	tailAngleSpeed float32 = 18.0
)

// distance returns the straight-line distance between two 2D points.
func distance(x1, y1, x2, y2 float32) float32 {
	deltaX := x1 - x2
	deltaY := y1 - y2
	return float32(math.Sqrt(float64(deltaX*deltaX + deltaY*deltaY)))
}

// approachFactor returns the share of the remaining gap to close this frame
// for a chase covering `speed` so-many times per second: 1 - exp(-speed*dt).
//
// The exponent can never reach 1, so the blob never overshoots its target
// however long a frame takes. The old dt*speed form is only the first-order
// approximation of this, and it passes 1 once a frame lasts longer than
// 1/speed, at which point the blob jumps past the target and wobbles.
// Expm1(-x) is used rather than 1-Exp(-x) so the value stays accurate when
// speed*dt is tiny, which is the usual case at 60fps.
func approachFactor(deltaTime, speed float32) float32 {
	return float32(-math.Expm1(-float64(deltaTime * speed)))
}

// approach moves current toward target, closing a fixed share of whatever gap
// is left every second. It never overshoots - the gap just keeps shrinking.
func approach(current, target, deltaTime, speed float32) float32 {
	return current + (target-current)*approachFactor(deltaTime, speed)
}

// tailAngle converts a screen-space direction into raylib rotation degrees.
// Screen axes grow down and to the right while raylib rotates clockwise from
// "up", so the direction needs its offset in y first and a quarter turn
// added at the end.
func tailAngle(deltaX, deltaY float32) float32 {
	return float32(math.Atan2(float64(deltaY), float64(deltaX)))*degreesPerRadian + 90
}

// dampAngle swings current toward target the short way around the circle, so
// the dome keeps rotating smoothly through the 0/360 seam instead of unwinding
// the long way round whenever the target angle wraps over.
//
// Used for the tail direction because the raw angle jumps by 180 degrees the
// moment the tail point crosses the head, and those jumps read as the clock
// hand snapping around rather than as the blob having changed its mind.
func dampAngle(current, target, deltaTime, speed float32) float32 {
	delta := target - current
	for delta > 180 {
		delta -= 360
	}
	for delta < -180 {
		delta += 360
	}
	return current + delta*approachFactor(deltaTime, speed)
}

// tailDimensions works out how to draw the stretched dome from the raw
// head-to-tail distance, returning its length and width in pixels.
//
// The width stays the blob's own width for the whole stretch, exactly like
// the original bball demo does it. That is what makes it read as a blob being
// pulled rather than a rubber band: an earlier version narrowed the dome as it
// lengthened, dividing the width by sqrt(stretch) to preserve volume, and it
// came out looking thin and pointed instead of like a fat blob.
//
// Only the length is capped, because the idle target teleports to wherever the
// last word was eaten and can therefore sit far away.
func tailDimensions(tailLength, blobWidth float32) (float32, float32) {
	if blobWidth <= 0 {
		return 0, 0
	}

	length := tailLength
	if length > maxTailStretch*blobWidth {
		length = maxTailStretch * blobWidth
	}

	// The width never changes: the dome is always as wide as the ball is.
	return length, blobWidth
}
