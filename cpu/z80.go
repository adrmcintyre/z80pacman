package cpu

import (
	"fmt"
	"sync/atomic"
)

// An IrqMode represents an interrupt handling mode.
type IrqMode int

const (
	// CPU executes instruction placed on bus by device, typically "rst n"
	IrqMode0 IrqMode = iota
	// CPU executes restart at 0038h
	IrqMode1
	// CPU calls 2-byte vector at memory[irqPage<<8 | bus & 0xfe]
	IrqMode2
)

// An Imm is a reference to the instruction stream
type Imm struct{}

var imm = Imm{}

// Implements the ByteRef interface
func (Imm) Rd() uint8 {
	return imm8()
}

// Implements the ByteRef interface
func (Imm) Wr(uint8) {
	panic("illegal write to immediate")
}

// Implements the WordRef interface
func (Imm) Rd16() uint16 {
	return imm16()
}

// Implements the WordRef interface
func (Imm) Wr16(uint16) {
	panic("illegal write to immediate")
}

// imm8 reads a byte from the instruction stream
func imm8() uint8 {
	loc := pc.Rd16()
	n := ref(loc).Rd()
	DebugTrace = append(DebugTrace, n)
	pc.Wr16(loc + 1)
	return n
}

// imm16 reads a word from the instruction stream (lo,hi)
func imm16() uint16 {
	nlo := imm8()
	nhi := imm8()
	return word(nhi, nlo)
}

type IndexMode int

const (
	IndexModeNone IndexMode = iota
	IndexModeIX
	IndexModeIY
)

// define machine state
var (
	// core registers
	b = new(Reg)
	c = new(Reg)
	d = new(Reg)
	e = new(Reg)
	h = new(Reg)
	l = new(Reg)
	a = new(Reg)
	f = new(Reg)

	// index register components
	ixh = new(Reg)
	ixl = new(Reg)
	iyh = new(Reg)
	iyl = new(Reg)

	// register pairs
	af = RegPair{a, f}
	bc = RegPair{b, c}
	de = RegPair{d, e}
	hl = RegPair{h, l}
	ix = RegPair{ixh, ixl}
	iy = RegPair{iyh, iyl}

	// dynamically adjusted during execution to point to hl, ix or iy
	hlMux = RegPair{h, l}

	pc = new(Reg16) // program counter
	sp = new(Reg16) // stack pointer

	// alternative register file for selected registers
	bc2 = new(Reg16)
	de2 = new(Reg16)
	hl2 = new(Reg16)
	af2 = new(Reg16)

	// irq enable flip-flops
	irqEnable1 = false // iff1
	irqEnable2 = false // iff2

	irqMode = IrqMode0 // selected interrupt mode
	irqPage uint8      // page used by IrqMode2

	indexMode IndexMode // has index prefix byte been seen?
	gotIndex  bool      // have we retrieved an index offset byte yet?
	index     uint8     // the index offset byte

	Halted         bool          // is processor currently halted, awaiting interrupt?
	DataBus        atomic.Uint32 // only lower 8 bits used
	AddressBus     atomic.Uint32 // only lower 16 bits used
	IrqAssertPin   atomic.Bool   // has an irq been asserted externally?
	ResetAssertPin atomic.Bool   // has reset been asserted externally?

	// for debugging
	DebugPC    uint16
	DebugTrace = make([]uint8, 4) // last 4 bytes retrieved from instruction stream
)

func Reset() {
	fmt.Println("RESET!")

	// reset all registers
	for _, r := range []WordRef{af, bc, de, hl, ix, iy, bc2, de2, hl2, af2, pc, sp} {
		r.Wr16(0)
	}

	// reset internal state
	Halted = false
	irqEnable1 = false
	irqEnable2 = false
	irqMode = IrqMode0
	irqPage = 0

	// reset outgoing signals
	DataBus.Store(0)
	AddressBus.Store(0)
}

func Step() uint16 {
	checkReset()
	checkInterrupt()

	DebugPC = pc.Rd16()
	if !Halted {
		nextInstruction()
	}
	return DebugPC
}

func nextInstruction() {
	DebugTrace = DebugTrace[:0]
	indexMode = IndexModeNone
	gotIndex = false
	hlMux = hl
	opcode := imm8()

	t := coreTable

	switch opcode {
	case 0xdd, 0xfd:
		if opcode == 0xdd {
			indexMode = IndexModeIX
			hlMux = ix
		} else {
			indexMode = IndexModeIY
			hlMux = iy
		}
		opcode = imm8()
		if opcode == 0xcb {
			index = imm8()
			gotIndex = true

			opcode = imm8()
			t = bitTable
		}
	case 0xcb:
		opcode = imm8()
		t = bitTable
	case 0xed:
		opcode = imm8()
		t = miscTable
	}

	t.fn[opcode]()
}

func checkReset() {
	if ResetAssertPin.Load() {
		//TODO
		//resetMachine()
	}
}

func checkInterrupt() {
	if irqEnable1 && IrqAssertPin.Load() {
		irqEnable1 = false
		irqEnable2 = false
		Halted = false
		IrqAssertPin.Store(false)
		switch irqMode {
		case IrqMode0:
			panic("irq mode 0 not implemented")
		case IrqMode1:
			call(0x38)
		case IrqMode2:
			low := uint8(DataBus.Load()) & 0xfe
			ind := word(irqPage, low)
			loc := ref(ind).Rd16()
			call(loc)
		}
	}
}
