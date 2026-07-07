package main

import "sync/atomic"

// Bits in dipswitch bank
const (
	// 00=2 coins, 1 play; 01=1 coin, 2 plays; 10=1 coin, 1 play; 11=free play
	DipCoins0 uint8 = 1 << iota
	DipCoins1

	// lives: 00=1 life, 01=2 lives, 10=3 lives, 11=5 lives
	DipLives0
	DipLives1

	// extra life at 00=10,000, 01=15,000, 10=20,000, 11=none
	DipBonus0
	DipBonus1

	LinkDifficulty // normal / hard (shorter blue/return times)
	LinkGhostNames // normal / alternative ghost names
)

var (
	dipSwitchState atomic.Uint32 // low 8 bits only
)

func dipsUpdate() {
	dipSwitchState.Store(mapOfBits{
		DipCoins0:      *flagCoins == 1 || *flagCoins == 2,
		DipCoins1:      *flagCoins == 2 || *flagCoins == 3,
		DipLives0:      *flagLives == 2 || *flagLives == 5,
		DipLives1:      *flagLives == 3 || *flagLives == 5,
		DipBonus0:      *flagBonus == 0 || *flagBonus == 15_000,
		DipBonus1:      *flagBonus == 0 || *flagBonus == 20_000,
		LinkDifficulty: !*flagDifficult,
		LinkGhostNames: !*flagAltGhosts,
	}.Uint32())
}
