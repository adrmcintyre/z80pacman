package main

import (
	"fmt"
	"os"

	"github.com/adrmcintyre/z80/cpu"
	"github.com/adrmcintyre/z80/video"
)

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
	for loc, cnt := range debugHisto {
		if cnt > 0 {
			fmt.Printf("%10d %04x %s\n", cnt, loc, cpu.Disassemble(uint16(loc), programROM[loc:loc+4]))
		}
	}
}

func dumpDisass() {
	fmt.Println("DISASSEMBLY")
	for loc, cnt := range debugHisto {
		if cnt > 0 {
			fmt.Printf("%04x %s\n", loc, cpu.Disassemble(uint16(loc), programROM[loc:loc+4]))
		}
	}
}

func dumpTrace(limit int) {
	fmt.Println("EXECUTION TRACE")
	n := len(debugTrace)
	if limit < 0 {
		limit = len(debugTrace)
	} else {
		limit = min(limit, n)
	}
	lim64 := min(uint64(limit), debugInstrCnt)

	i := int(debugInstrCnt%uint64(n) - lim64)
	for range lim64 {
		loc := debugTrace[(i+n)%n]
		fmt.Printf("%04x %s\n", loc, cpu.Disassemble(uint16(loc), programROM[loc:loc+4]))
		i++
	}
}

func dumpTraceLocs(limit int) {
	fmt.Println("PC TRACE")
	n := len(debugTrace)
	if limit < 0 {
		limit = len(debugTrace)
	} else {
		limit = min(limit, n)
	}
	i := int(debugInstrCnt%uint64(n)) - limit + 1
	j := 0
	for range limit {
		loc := debugTrace[(i+n)%n]
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
