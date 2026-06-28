package main

import (
	"fmt"
	"os"
)

var traceOn = true

func main() {
	const testMode = false
	if testMode {
		loadTestProgram()
	} else {
		load("pacman.rom")
		startVblankTicker()
	}

	powerOn()
	runCPU()
}

func resetMachine() {
	resetCPU()
	resetDevices()
}

func powerOn() {
	resetMachine()
	ioInit()
}

func resetDevices() {
	resetAssertPin.Store(false)
	irqAssertPin.Store(false)

	for i := range optionsRegister {
		optionsRegister[i].Store(false)
	}
	irqLowRegister.Store(0)
	watchDogRegister.Store(15)
}

func load(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open %s: %v\n", path, err)
		os.Exit(1)
	}
	copy(programROM[:], data)

	setBreakpoint(0x234b, Breakpoint{
		dumpTiles:   true,
		dumpProgram: true,
	})
}
