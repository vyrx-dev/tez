//go:build !test

package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func (g *Game) Draw() {
	if g.State == StateBoot {
		rl.ClearBackground(rl.Color{R: 0x11, G: 0x11, B: 0x11, A: 0xFF})

		alphaTez := float32(0)
		yOffsetTez := float32(20)

		if g.BootTimer < 0.5 {
			alphaTez = g.BootTimer / 0.5
			yOffsetTez = 20 * (1.0 - alphaTez)
		} else if g.BootTimer < 1.5 {
			alphaTez = 1.0
			yOffsetTez = 0
		} else {
			alphaTez = 1.0 - ((g.BootTimer - 1.5) / 0.5)
			yOffsetTez = 0
		}

		titleText := "Tez"
		titleSize := int32(120)
		titleWidth := rl.MeasureText(titleText, titleSize)

		rl.DrawText(titleText, int32(g.SW/2)-titleWidth/2, int32(g.SH/2)-(titleSize/2)+int32(yOffsetTez), titleSize, rl.Color{R: 255, G: 255, B: 255, A: uint8(alphaTez * 255)})

	} else if g.State == StatePlaying {
		rl.ClearBackground(rl.Color{R: 0x18, G: 0x18, B: 0x18, A: 0xFF})

		rl.BeginMode2D(g.Camera)

		hue := float32((g.Level * 40) % 360)
		blobColor := rl.ColorFromHSV(hue, 0.8, 1.0)

		textureWidth := float32(g.BallTexture.Width)
		textureHeight := float32(g.BallTexture.Height)
		displayWidth := textureWidth * g.BlobScale
		displayHeight := textureHeight * g.BlobScale

		tailLength, tailWidth := tailDimensions(g.TailLength, displayWidth)
		if tailLength > 1 {
			smearSourceRect := rl.Rectangle{X: 0, Y: 0, Width: textureWidth, Height: textureHeight * 0.5}
			smearDestRect := rl.Rectangle{X: g.BallX, Y: g.BallY, Width: tailWidth, Height: tailLength}
			rotationOrigin := rl.Vector2{X: tailWidth * 0.5, Y: tailLength}

			rl.DrawTexturePro(g.BallTexture, smearSourceRect, smearDestRect, rotationOrigin, g.TailAngle, blobColor)
		}

		ballSourceRect := rl.Rectangle{X: 0, Y: 0, Width: textureWidth, Height: textureHeight}
		ballDestRect := rl.Rectangle{
			X:      g.BallX - displayWidth*0.5,
			Y:      g.BallY - displayHeight*0.5,
			Width:  displayWidth,
			Height: displayHeight,
		}
		rl.DrawTexturePro(g.BallTexture, ballSourceRect, ballDestRect, rl.Vector2{}, 0, blobColor)

		for _, p := range g.Particles {
			alpha := uint8((p.Life / p.MaxLife) * 255)
			c := p.Color
			c.A = alpha
			rl.DrawCircle(int32(p.X), int32(p.Y), p.Size, c)
		}

		for _, t := range g.Targets {
			if t.Active {
				fontSize := int32(30 * t.Scale)
				textWidth := rl.MeasureText(t.Word, fontSize)
				startX := int32(t.X) - textWidth/2
				startY := int32(t.Y) - fontSize/2

				if t == g.ActiveTarget {
					rl.DrawText(t.Word, startX, startY, fontSize, rl.RayWhite)
				} else {
					rl.DrawText(t.Word, startX, startY, fontSize, rl.LightGray)
				}

				if t.TypedIndex > 0 {
					typedStr := t.Word[:t.TypedIndex]
					rl.DrawText(typedStr, startX, startY, fontSize, blobColor)
				}
			}
		}

		rl.EndMode2D()

		scoreText := fmt.Sprintf("LEVEL: %d | SCORE: %d | HIGH SCORE: %d", g.Level, g.Score, g.HighScore)
		rl.DrawText(scoreText, 10, 10, 20, rl.LightGray)

		for i := 0; i < g.Health; i++ {
			rl.DrawText("<3", int32(15+(i*40)), int32(g.SH-40), 25, rl.Red)
		}

		if g.LevelUpTimer > 0 {
			alpha := uint8((g.LevelUpTimer / 2.0) * 255)
			lvlText := fmt.Sprintf("LEVEL %d", g.Level)
			sz := int32(100)
			tw := rl.MeasureText(lvlText, sz)
			rl.DrawText(lvlText, int32(g.SW/2)-tw/2, int32(g.SH/2)-sz/2, sz, rl.Color{R: 255, G: 255, B: 255, A: alpha})
		}

	} else if g.State == StateGameOver {
		rl.ClearBackground(rl.Color{R: 0x05, G: 0x00, B: 0x00, A: 0xFF})

		for _, p := range g.Particles {
			alpha := uint8((p.Life / p.MaxLife) * 255)
			c := p.Color
			c.A = alpha
			rl.DrawCircle(int32(p.X), int32(p.Y), p.Size, c)
		}

		text := "GAME OVER"
		fontSize := int32(80)
		textWidth := rl.MeasureText(text, fontSize)
		rl.DrawText(text, int32(g.SW/2)-textWidth/2, int32(g.SH/2)-fontSize, fontSize, rl.Red)

		if g.NewPB {
			pbText := "NEW PERSONAL BEST!"
			pbSize := int32(50)
			pbWidth := rl.MeasureText(pbText, pbSize)
			rl.DrawText(pbText, int32(g.SW/2)-pbWidth/2, int32(g.SH/2)+10, pbSize, rl.Yellow)
		}

		subText := fmt.Sprintf("Final Score: %d", g.Score)
		subFontSize := int32(30)
		subTextWidth := rl.MeasureText(subText, subFontSize)
		yOffset := int32(g.SH/2) + 80
		if g.NewPB {
			yOffset += 30
		}
		rl.DrawText(subText, int32(g.SW/2)-subTextWidth/2, yOffset, subFontSize, rl.RayWhite)

		restartText := "Press ENTER to restart"
		resWidth := rl.MeasureText(restartText, 20)
		if int(rl.GetTime()*2)%2 == 0 {
			rl.DrawText(restartText, int32(g.SW/2)-resWidth/2, yOffset+60, 20, rl.LightGray)
		}
	}
}
