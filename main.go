//go:build !test

// Build:  nix-shell --run "go run ."
// Build (no raylib installed):  CGO_ENABLED=0 go run .
// Tests:  CGO_ENABLED=0 go test -tags test ./...
//
// The "test" tag keeps every file that touches raylib out of the test build,
// so the maths tests need no C compiler, no raylib and no display.

package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 600, "Tez")
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	rl.SetTargetFPS(60)

	game := NewGame()
	defer game.Close()

	for !rl.WindowShouldClose() {
		game.Update(rl.GetFrameTime())

		rl.BeginDrawing()
		game.Draw()
		rl.EndDrawing()
	}
}
