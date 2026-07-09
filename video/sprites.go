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
	spriteBpp    = 2
	spriteLen    = spriteWidth * spriteHeight * spriteBpp / 8
)

// spriteImages contains an ebiten Image for each sprite identifier.
var spriteImages [spriteCount]*ebiten.Image

func decodeSpriteData(rom []uint8) {
	// Decoding the sprites works similarly to decoding the tiles. Each byte
	// represents 4 pixels in a column, stored low bits then high bits just
	// like the tiles. Each 8 bytes draw a strip of 8x4 pixels, just like tiles.
	// These 8 strips are then arranged:
	// 5 1
	// 6 2
	// 7 3
	// 4 0
	//
	for spriteNum := range 64 {
		spriteImages[spriteNum] = ebiten.NewImage(spriteWidth, spriteHeight)
	}

	for i, b := range rom {
		spriteNum := i / spriteLen
		offset := i % spriteLen
		for pixelIndex := range 4 {
			x := 15 - ((offset>>2)&0x08 | offset&0x07)
			y := (((offset>>1)&0x0c | (3 - pixelIndex)) + 12) & 0x0f
			pHi := (b >> pixelIndex) & 1
			pLo := (b >> (pixelIndex + 4)) & 1
			spriteImages[spriteNum].Set(x, y, ColorChannel[pHi<<1|pLo])
		}
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

	if pal != 0 {
		colorm.DrawImage(img, spriteImages[look], ColorM[pal], &op)
	}
}
