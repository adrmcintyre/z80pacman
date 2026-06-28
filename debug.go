package main

import (
	"fmt"
	"os"
)

type Breakpoint struct {
	dumpTiles    bool
	dumpPalettes bool
	dumpProgram  bool
	resume       bool
	once         bool
}

func (bp Breakpoint) dump() {
	if bp.dumpTiles {
		dumpTileRAM()
	}
	if bp.dumpPalettes {
		dumpPaletteRAM()
	}
	if bp.dumpProgram {
		dumpProgramRAM()
	}
}

var breakpoints = make(map[uint16]Breakpoint, 0)

func setBreakpoint(loc uint16, bp Breakpoint) {
	breakpoints[loc] = bp
}

func breakpointed(prePC uint16) bool {
	if bp, ok := breakpoints[prePC]; ok {
		bp.dump()
		if bp.once {
			delete(breakpoints, prePC)
		}
		if !bp.resume {
			return true
		}
	}
	return false
}

func dumpTileRAM() {
	fmt.Println("TILES")
	wd := 32
	for i := range len(tileRAM) / wd {
		fmt.Printf("%04x |", 0x4000+i*wd)
		for j := range wd {
			fmt.Printf(" %02x", tileRAM[i*wd+j])
		}
		fmt.Printf("\n")
	}
	os.Exit(0)
}

func dumpPaletteRAM() {
	fmt.Println("PALETTES")
	wd := 32
	for i := range len(tileRAM) / wd {
		fmt.Printf("%04x |", 0x4400+i*wd)
		for j := range wd {
			fmt.Printf(" %02x", tileRAM[i*wd+j])
		}
		fmt.Printf("\n")
	}
	os.Exit(0)
}

func dumpProgramRAM() {
	wd := 16
	for i := range len(programRAM) / wd {
		fmt.Printf("%04x", 0x4c00+i*wd)
		for j := range wd / 8 {
			fmt.Printf(" |")
			for k := range 8 {
				fmt.Printf(" %02x", programRAM[i*wd+j*8+k])
			}
		}
		fmt.Printf("\n")
	}
	os.Exit(0)
}

func loadTestProgram() {
	org := uint16(0)
	asm := func(op uint8, args ...uint8) {
		programROM[org] = op
		org++
		for _, arg := range args {
			programROM[org] = arg
			org++
		}
	}

	asm(0xdd, 0x21, 0x00, 0x4c) // ld ix, 0x4c00
	//asm(0x01, 0x34, 0x12)       // ld bc, 0x1234
	//asm(0xed, 0x43, 0x00, 0x4c) // ld (0x4c00), bc
	//asm(0xed, 0x5b, 0x00, 0x4c) // ld de, (0x4c00)
	bp := org
	asm(0x76) // hlt

	breakpoints[bp] = Breakpoint{
		dumpProgram: true,
	}
}

func traceInstruction(prePC uint16) {
	hexBuf := ""
	for _, byt := range trace {
		hexBuf += fmt.Sprintf(" %02x", byt)
	}
	txtBuf := disass(prePC, trace)
	fmt.Printf("%04x |%-12s | %-16s | ", prePC, hexBuf, txtBuf)

	fmt.Printf("bc=%02x:%02x ", b.Rd(), c.Rd())
	fmt.Printf("de=%02x:%02x ", d.Rd(), e.Rd())
	fmt.Printf("hl=%02x:%02x ", h.Rd(), l.Rd())
	fmt.Printf("ix=%04x ", ix.Rd16())
	fmt.Printf("iy=%04x ", iy.Rd16())
	fmt.Printf("sp=%04x ", sp.Rd16())
	fmt.Printf("a=%02x ", a.Rd())
	flags := []byte("________")
	if flagS.get() {
		flags[7] = 'S'
	}
	if flagZ.get() {
		flags[6] = 'Z'
	}
	if flagH.get() {
		flags[4] = 'H'
	}
	if flagPV.get() {
		flags[2] = 'V'
	}
	if flagN.get() {
		flags[1] = 'N'
	}
	if flagC.get() {
		flags[0] = 'C'
	}
	fmt.Printf("f=%s\n", string(flags))
}

func disass(pc uint16, trace []uint8) string {
	switch trace[0] {
	case 0xdd, 0xfd:
		hlAlias := "hl"
		if trace[0] == 0xdd {
			hlAlias = "ix"
		} else {
			hlAlias = "iy"
		}
		if trace[1] == 0xcb {
			return bitTable.text(pc, trace[3], trace[2:], hlAlias)
		}
		return coreTable.text(pc, trace[1], trace[2:], hlAlias)

	case 0xed:
		return miscTable.text(pc, trace[1], trace[2:], "hl")
	case 0xcb:
		return bitTable.text(pc, trace[1], trace[2:], "hl")
	default:
		return coreTable.text(pc, trace[0], trace[1:], "hl")
	}
}
