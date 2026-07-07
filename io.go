package main

import (
	"fmt"
	"os"
)

func ioParseFlags() {
	switch *flagCoins {
	case 0, 1, 2, 3:
	default:
		fmt.Printf("illegal -dip-coins")
		os.Exit(1)
	}
	switch *flagLives {
	case 1, 2, 3, 5:
	default:
		fmt.Printf("illegal -dip-lives")
		os.Exit(1)
	}
	switch *flagBonus {
	case 0, 10_000, 15_000, 20_000:
	default:
		fmt.Printf("illegal -dip-bonus")
		os.Exit(1)
	}
}

func ioInit() {
	dipsUpdate()
	inputsUpdate()
}

// Helper for constructing input state.
type mapOfBits map[uint8]bool

func (mob mapOfBits) Uint32() (v uint32) {
	for i, bit := range mob {
		if bit {
			v |= uint32(i)
		}
	}
	return v
}
