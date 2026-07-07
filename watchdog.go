package main

import (
	"fmt"
	"sync/atomic"

	"github.com/adrmcintyre/z80pacman/z80"
)

var (
	watchdogRegister atomic.Uint32 // counts down from 15..0 (using top 4 bits)
)

func watchdogReset() {
	watchdogRegister.Store(15 << 28)
}

// Watchdog triggers RESET after 16 VBLANK if not cleared.
func watchdogTick() {
	if !*flagNoWatchdog {
		if watchdogRegister.Load() == 0 {
			fmt.Println("WATCHDOG!")
			z80.ResetAssertPin.Store(true)
		}
	}
	watchdogRegister.Add(15 << 28) // subtract 1 from top 4 bits
}
