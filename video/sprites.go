package video

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// A Sprite identifies a specific sprite bitmap.
type Sprite byte

const (
	spriteWidth  = 16
	spriteHeight = 16
	spriteCount  = 64
)

// A spriteBitmap encodes a 16x16, 2-bits-per-pixel sprite.
type spriteBitmap [spriteHeight]uint32

// spriteData defines a bitmap for each sprite identifier.
var spriteData = [spriteCount]spriteBitmap{
}

// spriteImages contains an ebiten Image for each sprite identifier.
var spriteImages [spriteCount]*ebiten.Image

// initSprites initialises the Image cache from the 2-bpp source data.
func initSprites() {
	for i, bitmap := range spriteData {
		img := ebiten.NewImage(spriteWidth, spriteHeight)
		for y, row := range bitmap {
			for x := range spriteWidth {
				img.Set(x, y, ColorChannel[row&0b11])
				row >>= 2
			}
		}
		spriteImages[i] = img
	}
}

// Draw paints the sprite onto img.
func (look Sprite) Draw(img *ebiten.Image, x, y int, flipX, flipY bool, pal Palette) {
	scaleX, scaleY := 1.0, 1.0
	if flipX {
		scaleX *= -1.0
		x += 16
	}
	if flipY {
		scaleY *= -1.0
		y += 16
	}
	// draw centred
	op := colorm.DrawImageOptions{}
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(float64(x), float64(y))

	// TODO - we don't seem to be processing sprite transparency correctly
	// (we seem to just render black) - just skip sprites with an all transparent
	// palette altogether for now.
	if pal != 0 {
		colorm.DrawImage(img, spriteImages[look], ColorM[pal], &op)
	}
}
