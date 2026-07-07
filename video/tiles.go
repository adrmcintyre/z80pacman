package video

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// A Tile identifies a specific tile bitmap.
type Tile byte

const (
	tileWidth  = 8
	tileHeight = 8
	tileCount  = 256
)

// A tileBitmap encodes an 8x8 2-bits-per-pixel tile image.
// The pixel order within each word is: 00:11:22:33:44:55:66:77
type tileBitmap [tileHeight]uint16

// tileData defines a bitmap for each tile identifier.
var tileData = [tileCount]tileBitmap{
}

// tileImages defines an ebiten Image for each tile identifier.
var tileImages [tileCount]*ebiten.Image

// initTiles initialises the Image cache for each tile.
func initTiles() {
	for i, bitmap := range tileData {
		img := ebiten.NewImage(tileWidth, tileHeight)
		for y, row := range bitmap {
			for x := range tileWidth {
				img.Set(x, y, ColorChannel[row&0b11])
				row >>= 2
			}
		}
		tileImages[i] = img
	}
}

// Draw paints the tile onto img.
func (t Tile) Draw(img *ebiten.Image, x, y int, pal Palette, flip bool) {
	op := colorm.DrawImageOptions{}
	if flip {
		op.GeoM.Scale(-1, -1)
		op.GeoM.Translate(float64(displayWidth-x), float64(displayHeight-y))
	} else {
		op.GeoM.Scale(1, 1)
		op.GeoM.Translate(float64(x), float64(y))
	}
	colorm.DrawImage(img, tileImages[t], ColorM[pal], &op)
}
