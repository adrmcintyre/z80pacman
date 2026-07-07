package main

import "time"

type Breakpoint struct {
	DumpTiles    bool // dump tile ram
	DumpPalettes bool // dump palette ram
	DumpProgram  bool // dump program state ram
	DumpHisto    bool // dump histogram of pc values
	DumpDisass   bool // disassemble executed locations
	DumpTrace    int  // disassemble last n executed instructions (0=all)
	Resume       bool // resume after breakpoint action
	Once         bool // disable breakpoint after first hit
}

var (
	debugBreakpoints = make(map[uint16]Breakpoint, 0)
)

func setBreakpoint(loc uint16, bp Breakpoint) {
	debugBreakpoints[loc] = bp
}

func breakpointed(pc uint16) bool {
	if bp, ok := debugBreakpoints[pc]; ok {
		bp.action()
		if bp.Once {
			delete(debugBreakpoints, pc)
		}
		if !bp.Resume {
			time.Sleep(1 * time.Second)
			return true
		}
	}
	return false
}

func (bp Breakpoint) action() {
	if bp.DumpTiles {
		dumpTileRAM()
	}
	if bp.DumpPalettes {
		dumpPaletteRAM()
	}
	if bp.DumpProgram {
		dumpProgramRAM()
	}
	if bp.DumpHisto {
		dumpHisto()
	}
	if bp.DumpDisass {
		dumpDisass()
	}
	if bp.DumpTrace != 0 {
		dumpTrace(bp.DumpTrace)
	}
}
