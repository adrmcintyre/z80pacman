package main

import "sync/atomic"

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
	n := mem(loc).Rd()
	trace = append(trace, n)
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
	b = new(Reg)
	c = new(Reg)
	d = new(Reg)
	e = new(Reg)
	h = new(Reg)
	l = new(Reg)
	a = new(Reg)
	f = new(Reg)

	ixh = new(Reg)
	ixl = new(Reg)
	iyh = new(Reg)
	iyl = new(Reg)

	af    = RegPair{a, f}
	bc    = RegPair{b, c}
	de    = RegPair{d, e}
	hl    = RegPair{h, l}
	ix    = RegPair{ixh, ixl}
	iy    = RegPair{iyh, iyl}
	hlMux = RegPair{h, l}

	bc2 = new(Reg16)
	de2 = new(Reg16)
	hl2 = new(Reg16)
	af2 = new(Reg16)

	pc = new(Reg16)
	sp = new(Reg16)

	trace = make([]uint8, 4)

	halted     = false
	irqEnable1 = false // iff1
	irqEnable2 = false // iff2
	irqMode    = IrqMode0
	irqPage    uint8

	indexMode IndexMode
	gotIndex  bool
	index     uint8

	dataBus        atomic.Uint32 // only lower 8 bits used
	addressBus     atomic.Uint32 // only lower 16 bits used
	irqAssertPin   atomic.Bool
	resetAssertPin atomic.Bool
)

func resetCPU() {
	// reset all registers
	for _, r := range []WordRef{af, bc, de, hl, ix, iy, bc2, de2, hl2, af2, pc, sp} {
		r.Wr16(0)
	}

	// reset internal state
	halted = false
	irqEnable1 = false
	irqEnable2 = false
	irqMode = IrqMode0
	irqPage = 0

	// reset outgoing signals
	dataBus.Store(0)
	addressBus.Store(0)
}

func runCPU() {
	for {
		if resetAssertPin.Load() {
			resetMachine()
		}

		if irqEnable1 && irqAssertPin.Load() {
			halted = false
			irqAssertPin.Store(false)
			switch irqMode {
			case IrqMode0:
				panic("irq mode 0 not implemented")
			case IrqMode1:
				call(0x38)
			case IrqMode2:
				low := uint8(dataBus.Load()) & 0xfe
				loc := mem(word(irqPage, low)).Rd16()
				call(loc)
			}
		}

		if halted {
			continue
		}

		prePC := pc.Rd16()

		if stop := breakpointed(prePC); stop {
			break
		}

		trace = trace[:0]
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

		if traceOn {
			traceInstruction(prePC)
		}
	}
}
