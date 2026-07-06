package video

import (
	"image/color"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	hOffset int = 8
	vOffset int = 8
)

// Video abstracts the video hardware.
var (
	Mutex   sync.Mutex    // controls writes
	TileRAM [1024]Tile    // tiles
	PalRAM  [1024]Palette // per-tile colour palettes

	// layout:
	// +0 sprite 0: [7:2]=look, [1]=flipx, [0]=flipy
	// +1 sprite 0: [7:0]=color
	// +2 sprite 1: ...
	// +f ...
	SpriteLookRAM [0x010]uint8

	// layout:
	// +0 sprite 0: X
	// +1 sprite 0: Y
	// +2 sprite 1: ...
	// +f ...
	SpritePosRegister [16]uint8

	//	spriteLookRAM     [16]uint8
	//	spritePosRegister [16]uint8
	_ *ebiten.Shader // shader for output filtering
)

var (
	flipScreen  atomic.Bool
	player1Lamp atomic.Bool
	player2Lamp atomic.Bool
)

// tileIndex converts tile co-ordinates to an index into
// tileRam / paletteRam.
//
//	top (0 <= x < 32, y < 2) - note x=0,1,30,31 are invisible
//	index := (y+30)*32 + (31-x)				// 0x3c0-0x3ff
//
//	normal (0 <= x < 28, 2 <= y < 34)
//	index := (29-x)*32 + (y - 2) 	// 0x040-0x3bf
//
//	bottom (0 <= x < 32, y < 2) - note x=0,1,30,31 are invisible
//	index := y*32 + (31-x) 					// 0x000-0x03f
func tileIndex(x int, y int) int {
	switch {
	case y < 2:
		return 0x3c0 + y*32 + 31 - x
	case y >= 34:
		return 0x000 + (y-34)*32 + 31 - x
	default:
		return 0x40 + (27-x)*32 + (y - 2)
	}
}

// DrawTiles paints the supplied bitmap with the contents
// of tile ram mixed with the colours from palette ram.
func drawTiles(screen *ebiten.Image) {
	for ty := range 36 {
		for tx := range 28 {
			posX, posY := tx*8, ty*8
			index := tileIndex(tx, ty)
			t := TileRAM[index]
			pal := PalRAM[index] & 0x3f
			t.Draw(screen, hOffset+posX, vOffset+posY, pal)
		}
	}
}

func drawSprites(screen *ebiten.Image) {
	// note, sprites 0 and 7 are never used by pacman
	for i := 0; i <= 7; i++ {
		x := int(SpritePosRegister[2*i])
		y := int(SpritePosRegister[2*i+1])
		x = 224 - x + 24 // TODO perhaps tiles need the opposite offset, or some combo?
		y = 288 - y - 8
		data := SpriteLookRAM[2*i]
		pal := Palette(SpriteLookRAM[2*i+1] & 0b0011_1111)
		look := Sprite(data >> 2)
		flipX := data&0b0000_0010 != 0
		flipY := data&0b0000_0001 != 0
		look.Draw(screen, x, y, flipX, flipY, pal)
	}
}

// Draw paints the supplied bitmap with tiles, with all sprites
// established for this frame rendered on top.
func Draw(screen *ebiten.Image) {
	Mutex.Lock()
	defer Mutex.Unlock()
	drawTiles(screen)
	drawSprites(screen)

	// TODO display lamps in a less obnoxious way
	if player1Lamp.Load() {
		vector.FillCircle(screen, 8, 8, 4, color.White, false)
	}
	if player2Lamp.Load() {
		vector.FillCircle(screen, 232, 8, 4, color.White, false)
	}
}

func SetFlipScreen(value bool) {
	flipScreen.Store(value)
}

func SetPlayer1Lamp(value bool) {
	player1Lamp.Store(value)
}

func SetPlayer2Lamp(value bool) {
	player2Lamp.Store(value)
}
