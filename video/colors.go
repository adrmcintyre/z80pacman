package video

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// A colorByte is a bb:ggg:rrr colour triplet.
type colorByte byte

// Colour palettes contain colour indexes corresponding to entries in this table.
// We use octal notation below as it nicely breaks out each colour channel.
var colorData = [32]colorByte{
}

// ColorChannel maps each possible value in a 2-bpp bitmap to a single colour channel.
var ColorChannel = []color.Color{
	color.RGBA{},                       // transparent
	color.RGBA{0x00, 0xff, 0x00, 0xff}, // green
	color.RGBA{0xff, 0x00, 0x00, 0xff}, // red
	color.RGBA{0x00, 0x00, 0xff, 0xff}, // blue
}

// toRGB converts a colorByte to float64 RGB.
// The result is written to the provided pointers.
func (bbgggrrr colorByte) toRGB() (float64, float64, float64) {
	bb := (bbgggrrr >> 6) & 0b11
	bbb := byte(bb << 1)
	ggg := byte((bbgggrrr >> 3) & 0b111)
	rrr := byte((bbgggrrr >> 0) & 0b111)

	return resistorLadder(rrr), resistorLadder(ggg), resistorLadder(bbb)
}

// resistorLadder simulates the resistor resistorLadder used for converting
// the asserted logic levels to currents to drive the display
// circuitry. The result is the proportion of maximum drive
// current, between 0.0 and 1.0
func resistorLadder(bits byte) float64 {
	// currents from 5V across 220, 470 and 1k ohm resistors
	const high = 5.0 / 220
	const mid = 5.0 / 470
	const low = 5.0 / 1000
	const limit = (high + mid + low)
	current := 0.0000
	if bits&(1<<2) != 0 {
		current += high
	}
	if bits&(1<<1) != 0 {
		current += mid
	}
	if bits&(1<<0) != 0 {
		current += low
	}
	return current / limit
}

var (
	// ColorM contains an ebiten colour matrix corresponding to each paletteEntry.
	ColorM [64]colorm.ColorM
)

// InitColors initialises the cache of ebiten.ColorM matrices corresponding to each
// colour palette.
//
// Fortuitously a colour matrix represents four colour channels R, G, B and A.
// We assign the 3 palette entries for 10, 01 and 11 to channels R, G and B
// respectively, and 00 to A. We then set up the matrix so that 100% R will
// render as the color corresponding to 10's color index, and similarly with
// G for 01 and B for 11. When creating full-colour RGBA ebiten bitmaps from
// the 2-bpp sources, we replace colour 10 with full-scale red, 01 with green,
// and 11 with blue. Finally when compositing the bitmaps into a display frame,
// we mix them with the colour matrix corresponding to the selected palette,
// and R, G and B are interpreted as the desired colours.
func InitColors() {
	for i := range 64 {
		mat := colorm.ColorM{}
		for j := range 3 {
			ci := paletteData[i][1+j]
			r, g, b := colorData[ci].toRGB()
			mat.SetElement(0, j, r)
			mat.SetElement(1, j, g)
			mat.SetElement(2, j, b)
		}
		ColorM[i] = mat
	}
}
