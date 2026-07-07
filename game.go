package main

import (
	"github.com/adrmcintyre/z80pacman/video"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct{}

func (g *Game) Update() error {
	inputsUpdate()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	video.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return video.Layout(outsideWidth, outsideHeight)
}
