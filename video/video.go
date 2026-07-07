package video

import (
	"image/color"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	hTiles, vTiles = 28, 36 // dimensions of display area in tiles
	displayWidth   = hTiles * tileWidth
	displayHeight  = vTiles * tileHeight
	borderWidth    = 4
)

const (
	hOffset = borderWidth
	vOffset = borderWidth
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

	flipScreen  atomic.Bool
	player1Lamp atomic.Bool
	player2Lamp atomic.Bool

	frameBuf *ebiten.Image
)

var (
	_ *ebiten.Shader // shader for output filtering
)

func Init() {
	initColors()
	initTiles()
	initSprites()
	initEbiten(28.0/36.0, 0.75)
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

func initEbiten(aspectRatio float64, fillRatio float64) {
	ebiten.SetWindowTitle("z80-pacman")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	w, h := ebiten.Monitor().Size()
	fw, fh := float64(w), float64(h)
	if fw/fh > aspectRatio {
		w = int(fh * aspectRatio)
	} else {
		h = int(fw / aspectRatio)
	}
	ebiten.SetWindowSize(int(float64(w)*fillRatio), int(float64(h)*fillRatio))
	frameBuf = ebiten.NewImage(displayWidth, displayHeight)
}

// Draw paints the supplied bitmap with tiles, with all sprites
// established for this frame rendered on top.
func Draw(screen *ebiten.Image) {
	Mutex.Lock()
	defer Mutex.Unlock()

	frameBuf.Clear()
	drawTiles(frameBuf)
	drawSprites(frameBuf)

	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, 1)
	op.GeoM.Translate(float64(hOffset), float64(vOffset))
	screen.DrawImage(frameBuf, &op)

	// TODO display lamps in a less obnoxious way
	if player1Lamp.Load() {
		vector.FillCircle(screen, 8, 8, 4, color.White, false)
	}
	if player2Lamp.Load() {
		vector.FillCircle(screen, hOffset+224+borderWidth-8, 8, 4, color.White, false)
	}
}

func Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	const (
		// calculate logical output size
		logicalWidth  = float64(displayWidth + 2*borderWidth)
		logicalHeight = float64(displayHeight + 2*borderWidth)
		logicalAspect = float64(logicalWidth) / float64(logicalHeight)
	)

	var (
		fOutsideWidth  = float64(outsideWidth)
		fOutsideHeight = float64(outsideHeight)
		outputAspect   = fOutsideWidth / fOutsideHeight

		fScreenWidth  = logicalWidth
		fScreenHeight = logicalHeight
	)

	// centre output in window
	if outputAspect > logicalAspect {
		fScreenWidth = logicalHeight * outputAspect
	} else {
		fScreenHeight = logicalWidth / outputAspect
	}
	return int(fScreenWidth), int(fScreenHeight)
}

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

// drawTiles paints the supplied bitmap with the contents
// of tile ram mixed with the colours from palette ram.
func drawTiles(img *ebiten.Image) {
	flip := flipScreen.Load()
	for ty := range vTiles {
		for tx := range hTiles {
			posX, posY := tx*tileWidth, ty*tileHeight
			index := tileIndex(tx, ty)
			t := TileRAM[index]
			pal := PalRAM[index] & 0x3f
			t.Draw(img, posX, posY, pal, flip)
		}
	}
}

func drawSprites(img *ebiten.Image) {
	// note, sprites 0 and 7 are never used by pacman
	for i := 0; i <= 7; i++ {
		x := int(SpritePosRegister[2*i])
		y := int(SpritePosRegister[2*i+1])
		x = displayWidth - x + 16
		y = displayHeight - y - 16
		data := SpriteLookRAM[2*i]
		pal := Palette(SpriteLookRAM[2*i+1] & 0b0011_1111)
		look := Sprite(data >> 2)
		flipX := data&0b0000_0010 != 0
		flipY := data&0b0000_0001 != 0
		look.Draw(img, x, y, flipX, flipY, pal)
	}
}
