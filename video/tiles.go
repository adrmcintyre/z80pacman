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
	tileBpp    = 2
	tileLen    = tileWidth * tileHeight * tileBpp / 8
)

// tileImages defines an ebiten Image for each tile identifier.
var tileImages [tileCount]*ebiten.Image

func decocdeTileData(rom []byte) {
	// The 4 pixels from each byte are vertical strips.
	// The first 8 bytes draw the lower half of the tile, left to right, top to bottom.
	// The last 8 bytes defining the tile draw the top half of the tile.
	//
	// We decode this into a saner format.
	for tileNum := range 256 {
		tileImages[tileNum] = ebiten.NewImage(tileWidth, tileHeight)
	}
	for i, b := range rom {
		tileNum := i / tileLen
		offset := i % tileLen
		for pixelIndex := range 4 {
			x := 7 - (offset & 0x07)
			y := (((offset>>1)&0x04 | (3 - pixelIndex)) + 4) & 0x07
			pHi := (b >> pixelIndex) & 1
			pLo := (b >> (pixelIndex + 4)) & 1
			tileImages[tileNum].Set(x, y, ColorChannel[pHi<<1|pLo])
		}
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
