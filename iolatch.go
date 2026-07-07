package main

import (
	"sync/atomic"

	"github.com/adrmcintyre/z80pacman/audio"
	"github.com/adrmcintyre/z80pacman/video"
)

// An 8-way latch the software can write to to control the hardware.
const (
	LatchIrqEnable      = iota // 1 enable, 0 disable
	LatchSoundEnable           // 1 enable, 0 disable
	LatchAuxBoardEnable        // not connected
	LatchFlipScreen            // 0 normal, 1 flipped for cocktail
	LatchPlayer1Lamp           //
	LatchPlayer2Lamp           //
	LatchCoinLockout           //
	LatchCoinCounter           //
)

var (
	irqLowRegister    atomic.Uint32 // lower 8 bits only
	irqEnableRegister atomic.Bool
)

func latchWrite(addr int, value bool) {
	switch addr {
	case LatchIrqEnable:
		irqEnableRegister.Store(value)
	case LatchSoundEnable:
		audio.SetSoundEnable(value)
	case LatchAuxBoardEnable:
		// do nothing
	case LatchFlipScreen:
		video.SetFlipScreen(value)
	case LatchPlayer1Lamp:
		video.SetPlayer1Lamp(value)
	case LatchPlayer2Lamp:
		video.SetPlayer2Lamp(value)
	case LatchCoinLockout:
		// do nothing
	case LatchCoinCounter:
		// do nothing
	}
}
