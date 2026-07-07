package main

import (
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

// Bits in in0 bank
const (
	In0_NotUp uint8 = 1 << iota
	In0_NotLeft
	In0_NotRight
	In0_NotDown
	In0_NotRackTest // (go to next level)
	In0_RisingCoin1 // trigger by going from 0 to 1
	In0_RisingCoin2 // trigger by going from 0 to 1
	In0_NotCredit
)

// Bits in in1 bank
const (
	In1_NotUp uint8 = 1 << iota
	In1_NotLeft
	In1_NotRight
	In1_NotDown
	In1_NotTest
	In1_NotStart1
	In1_NotStart2
	In1_NotCocktailMode
)

var (
	in0_State atomic.Uint32 // low 8 bits only
	in1_State atomic.Uint32 // low 8 bits only
)

func inputsUpdate() {
	in0_State.Store(mapOfBits{
		In0_NotUp:       !ebiten.IsKeyPressed(ebiten.KeyUp),
		In0_NotLeft:     !ebiten.IsKeyPressed(ebiten.KeyLeft),
		In0_NotRight:    !ebiten.IsKeyPressed(ebiten.KeyRight),
		In0_NotDown:     !ebiten.IsKeyPressed(ebiten.KeyDown),
		In0_NotRackTest: !*flagRackTest,
		In0_RisingCoin1: !false,
		In0_RisingCoin2: !false,
		In0_NotCredit:   !ebiten.IsKeyPressed(ebiten.KeyC),
	}.Uint32())

	// use same arrow keys for both players
	in1_State.Store(mapOfBits{
		In1_NotUp:           !ebiten.IsKeyPressed(ebiten.KeyUp),
		In1_NotLeft:         !ebiten.IsKeyPressed(ebiten.KeyLeft),
		In1_NotRight:        !ebiten.IsKeyPressed(ebiten.KeyRight),
		In1_NotDown:         !ebiten.IsKeyPressed(ebiten.KeyDown),
		In1_NotTest:         (!*flagTest) != ebiten.IsKeyPressed(ebiten.KeyT),
		In1_NotStart1:       !ebiten.IsKeyPressed(ebiten.Key1),
		In1_NotStart2:       !ebiten.IsKeyPressed(ebiten.Key2),
		In1_NotCocktailMode: !*flagCocktail,
	}.Uint32())
}
