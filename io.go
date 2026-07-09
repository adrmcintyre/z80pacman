package main

func ioParseFlags() {
	switch *flagCoins {
	case 0, 1, 2, 3:
	default:
		die("illegal -dip-coins")
	}
	switch *flagLives {
	case 1, 2, 3, 5:
	default:
		die("illegal -dip-lives")
	}
	switch *flagBonus {
	case 0, 10_000, 15_000, 20_000:
	default:
		die("illegal -dip-bonus")
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
