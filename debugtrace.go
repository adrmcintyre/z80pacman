package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adrmcintyre/z80pacman/z80"
)

var (
	debugMem      [0x10000]bool
	debugHisto    [0x10000]int64 // histogram indexed by pc for each instruction executed
	debugTrace    [1024]uint16   // last 1024 values of pc (circular buffer)
	debugInstrCnt uint64         // increments after every instruction executed
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
			debugMem[i] = true
		}
	}
}

func debugTraceCPU(pc uint16) {
	if !z80.Halted {
		debugHisto[pc]++
		debugTrace[debugInstrCnt%uint64(len(debugTrace))] = pc
		debugInstrCnt++
		if *flagTrace {
			debugTraceInstruction(pc, z80.DebugTrace)
		}
	}
}

func debugTraceInstruction(pc uint16, trace []byte) {
	hexBuf := ""
	for _, byt := range trace {
		hexBuf += fmt.Sprintf(" %02x", byt)
	}
	txtBuf := z80.Disassemble(pc, trace)
	fmt.Printf("%04x |%-12s | %-16s | ", pc, hexBuf, txtBuf)

	s := z80.GetState()

	fmt.Printf("a=%02x ", s.A)
	fmt.Printf("bc=%02x:%02x ", s.B, s.C)
	fmt.Printf("de=%02x:%02x ", s.D, s.E)
	fmt.Printf("hl=%02x:%02x ", s.H, s.L)
	fmt.Printf("ix=%04x ", s.IX)
	fmt.Printf("iy=%04x ", s.IY)
	fmt.Printf("sp=%04x ", s.SP)
	flags := []byte("________")
	if s.FS {
		flags[7] = 'S'
	}
	if s.FZ {
		flags[6] = 'Z'
	}
	if s.FH {
		flags[4] = 'H'
	}
	if s.FPV {
		flags[2] = 'V'
	}
	if s.FN {
		flags[1] = 'N'
	}
	if s.FC {
		flags[0] = 'C'
	}
	fmt.Printf("f=%s\n", string(flags))
}

func debugTraceRead(addr uint16, data uint8) {
	if debugMem[addr] {
		fmt.Printf("%04x READ %02x\n", addr, data)
	}
}

func debugTraceWrite(addr uint16, data uint8) {
	if debugMem[addr] {
		fmt.Printf("%04x WRITE %02x\n", addr, data)
	}
}
