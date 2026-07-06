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

func debugInit() {
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
	dumpTiles    bool
	dumpPalettes bool
	dumpProgram  bool
	dumpHisto    bool
	dumpDisass   bool
	dumpTrace    int
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

var org uint16

func asm(op uint8, args ...uint8) {
	programROM[org] = op
	org++
	for _, arg := range args {
		programROM[org] = arg
		org++
	}
}
func rel(dst uint16) uint8 {
	return uint8(dst - (org + 2))
}

func loadTestProgram(prog string) {
	asm(0xf3)             // di
	asm(0x3e, 0x00)       // ld a, #00
	asm(0xed, 0x47)       // ld i, a
	asm(0xc3, 0x00, 0x01) // jp #0100
	org = 0x100

	switch prog {
	case "1":
		test1()
	case "2":
		test2()
	}

	bp := org
	asm(0x76) // hlt

	breakpoints[bp] = Breakpoint{resume: false, dumpTiles: true, dumpPalettes: true}
}

func test1() {
	freq0 := (440 * 4096 / 375)
	freq1 := (523 * 4096 / 375)
	freq2 := (660 * 4096 / 375)

	// channel 0
	asm(0xdd, 0x21, 0x40, 0x50) // ld ix, #5040 // acc/wave, followed by freq/vol

	asm(0xdd, 0x36, 0x10, uint8(freq0>>0))  // ld (ix+off), #val
	asm(0xdd, 0x36, 0x11, uint8(freq0>>4))  // ld (ix+off), #val
	asm(0xdd, 0x36, 0x12, uint8(freq0>>8))  // ld (ix+off), #val
	asm(0xdd, 0x36, 0x13, uint8(freq0>>12)) // ld (ix+off), #val
	asm(0xdd, 0x36, 0x14, uint8(freq0>>16)) // ld (ix+off), #val
	asm(0xdd, 0x36, 0x15, 0x0f)             // max volume

	// channel 1
	asm(0xdd, 0x36, 0x16, uint8(freq1>>4))  // freq
	asm(0xdd, 0x36, 0x17, uint8(freq1>>8))  // freq
	asm(0xdd, 0x36, 0x18, uint8(freq1>>12)) // freq
	asm(0xdd, 0x36, 0x19, uint8(freq1>>16)) // freq
	asm(0xdd, 0x36, 0x1a, 0x0f)             // max volume

	// channel 2
	asm(0xdd, 0x36, 0x0f, 0x00)             // voice
	asm(0xdd, 0x36, 0x1b, uint8(freq2>>4))  // freq
	asm(0xdd, 0x36, 0x1c, uint8(freq2>>8))  // freq
	asm(0xdd, 0x36, 0x1d, uint8(freq2>>12)) // freq
	asm(0xdd, 0x36, 0x1e, uint8(freq2>>16)) // freq
	asm(0xdd, 0x36, 0x1f, 0x0f)             // max volume
}

func test2() {
	// labels
	var (
		next_freq uint16
		got_frame uint16
		halt_lp   uint16
	)

	const (
		irq_vec       uint16 = 0x00fe
		prog_start    uint16 = 0x1000
		stack_top     uint16 = 0x4fc0
		io_irq_en     uint16 = 0x5000
		io_audio_base uint16 = 0x5040
	)

	const (
		freq0  uint16 = 440 * 4096 / 375
		delta0 uint16 = 100

		nth_frame uint8 = 5
		cnt       uint8 = 40
	)

	asm(0x31, lo(stack_top), hi(stack_top)) // ld sp, #stack_top
	asm(0xed, 0x5e)                         // im 2 - interrupt mode 2
	asm(0x3e, hi(irq_vec))                  // ld a, #00
	asm(0xed, 0x47)                         // ld i, a // vector via 00..
	asm(0x3e, lo(irq_vec))                  // ld a, #fe
	asm(0xd3, 0x00)                         // out (0), a	// vector via 00fe
	asm(0x3e, 0x01)                         // ld a, #01
	asm(0x32, lo(io_irq_en), hi(io_irq_en)) // ld (#5000),a ; enable interrupt hardware

	asm(0x21, lo(freq0), hi(freq0))   // ld hl, #freq0
	asm(0x11, lo(delta0), hi(delta0)) // ld de, #delta0
	asm(0x0e, nth_frame)              // ld c, #nth_frame
	asm(0x06, cnt)                    // ld b, #cnt
	asm(0xfb)                         // ei ; enable interrupts

	halt_lp = org
	asm(0x76)               // hlt ; wait for interrupt
	asm(0x18, rel(halt_lp)) // jr halt_lp

	org = irq_vec
	asm(lo(prog_start), hi(prog_start)) // point interrupt vector at 0x1000

	for range 2 {
		org = prog_start

		asm(0xdd, 0x21, lo(io_audio_base), hi(io_audio_base)) // ld ix, #5040 // acc/wave, followed by freq/vol

		asm(0x0d)                 // dec c
		asm(0x28, rel(got_frame)) // jr z, ...
		asm(0xfb)                 // ei
		asm(0xc9)                 // ret

		got_frame = org
		asm(0x0e, nth_frame)            // ld c, #nth_frame
		asm(0x05)                       // dec b
		asm(0x20, rel(next_freq))       // jr nz, next_freq
		asm(0x21, lo(freq0), hi(freq0)) // ld hl, #freq0
		asm(0x06, cnt)                  // ld b, #cnt

		next_freq = org
		asm(0x19)                   // add hl, de
		asm(0x7d)                   // ld a,l
		asm(0xdd, 0x77, 0x10)       // ld (ix+#10), a
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0xdd, 0x77, 0x11)       // ld (ix+#11), a
		asm(0x7c)                   // ld a,h
		asm(0xdd, 0x77, 0x12)       // ld (ix+#12), a
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0xdd, 0x77, 0x13)       // ld (ix+#13), a
		asm(0xdd, 0x36, 0x15, 0x0f) // max volume

		asm(0xfb) // ei
		asm(0xc9) // ret
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
