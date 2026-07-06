package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adrmcintyre/z80/video"
)

var (
	traceMem [0x10000]bool
)

func debugParseFlags() {
	ftm := *flagTraceMem
	if ftm == "" {
		return
	}
	for _, part := range strings.Split(ftm, ",") {
		start, end, ok := strings.Cut(part, "-")
		if !ok {
			start = part
			end = part
		}
		s, err := strconv.ParseInt(start, 16, 64)
		if err != nil || s < 0 || s > 0xffff {
			fmt.Printf("bad memory range specified: %s", ftm)
			os.Exit(1)
		}
		e, err := strconv.ParseInt(end, 16, 64)
		if err != nil || e < s || e > 0xffff {
			fmt.Printf("bad memory range specified: %s\n", ftm)
			os.Exit(1)
		}
		for i := s; i <= e; i++ {
			traceMem[i] = true
		}
	}
}

type Breakpoint struct {
	dumpTiles    bool // dump tile ram
	dumpPalettes bool // dump palette ram
	dumpProgram  bool // dump program state ram
	dumpHisto    bool // dump histogram of pc values
	dumpDisass   bool // disassemble executed locations
	dumpTrace    int  // disassemble last n executed instructions (0=all)
	resume       bool // resume after breakpoint action
	once         bool // disable breakpoint after first hit
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
	if bp.dumpHisto {
		dumpHisto()
	}
	if bp.dumpDisass {
		dumpDisass()
	}
	if bp.dumpTrace != 0 {
		dumpTrace(bp.dumpTrace)
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
			time.Sleep(10 * time.Second)
			return true
		}
	}
	return false
}

func dumpTileRAM() {
	fmt.Println("TILE RAM")
	wd := 32
	for i := range len(video.TileRAM) / wd {
		fmt.Printf("%04x |", 0x4000+i*wd)
		for j := range wd {
			fmt.Printf(" %02x", video.TileRAM[i*wd+j])
		}
		fmt.Printf("\n")
	}
}

func dumpPaletteRAM() {
	fmt.Println("PALETTE RAM")
	wd := 32
	for i := range len(video.PalRAM) / wd {
		fmt.Printf("%04x |", 0x4400+i*wd)
		for j := range wd {
			fmt.Printf(" %02x", video.PalRAM[i*wd+j])
		}
		fmt.Printf("\n")
	}
}

func dumpProgramRAM() {
	fmt.Println("PROGRAM RAM")
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

func dumpHisto() {
	fmt.Println("PC HISTOGRAM")
	for loc, cnt := range pcHisto {
		if cnt > 0 {
			fmt.Printf("%10d %04x %s\n", cnt, loc, disass(uint16(loc), programROM[loc:loc+4]))
		}
	}
}

func dumpDisass() {
	fmt.Println("DISASSEMBLY")
	for loc, cnt := range pcHisto {
		if cnt > 0 {
			fmt.Printf("%04x %s\n", loc, disass(uint16(loc), programROM[loc:loc+4]))
		}
	}
}

func dumpTrace(limit int) {
	fmt.Println("EXECUTION TRACE")
	n := len(pcTrace)
	if limit < 0 {
		limit = len(pcTrace)
	} else {
		limit = min(limit, n)
	}
	i := pcTraceIndex - limit + 1
	for range limit {
		loc := pcTrace[(i+n)%n]
		fmt.Printf("%04x %s\n", loc, disass(uint16(loc), programROM[loc:loc+4]))
		i++
	}
}

func dumpTraceLocs(limit int) {
	fmt.Println("PC TRACE")
	n := len(pcTrace)
	if limit < 0 {
		limit = len(pcTrace)
	} else {
		limit = min(limit, n)
	}
	i := pcTraceIndex - limit + 1
	j := 0
	for range limit {
		loc := pcTrace[(i+n)%n]
		fmt.Printf("%04x ", loc)
		i++
		j++
		if j%8 == 0 {
			fmt.Println()
		}
	}
	if j%8 != 0 {
		fmt.Println()
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
