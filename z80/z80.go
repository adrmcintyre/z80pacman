package z80

import (
	"sync/atomic"
)

// An IrqMode represents an interrupt handling mode.
type IrqMode int

const (
	// CPU executes instruction placed on bus by device, typically "rst n"
	IrqMode0 IrqMode = iota
	// CPU executes restart at 0038h
	IrqMode1
	// CPU calls 2-byte vector at memory[irqPage<<8 | irqData & 0xfe]
	IrqMode2
)

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

	resetPending bool       // has Reset been called?
	nmiPending   bool       // has NMI been called?
	irqPending   bool       // has IRQ been called?
	irqData      uint8      // state of data bus supplied when IRQ was called
	irqMode      = IrqMode0 // selected interrupt mode
	irqPage      uint8      // page used by IrqMode2

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
	// reset all registers
	for _, r := range []WordRef{af, bc, de, hl, ix, iy, bc2, de2, hl2, af2, pc, sp} {
		r.Wr16(0)
	}

	// reset internal state
	resetPending = false
	nmiPending = false
	irqPending = false
	irqData = 0
	irqEnable1 = false
	irqEnable2 = false
	irqMode = IrqMode0
	irqPage = 0

	Halted = false
}

func NMI() {
	nmiPending = true
}

func IRQ(data uint8) {
	irqPending = true
	irqData = data
}

func Step() uint16 {
	checkSignals()

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

func checkSignals() {
	if resetPending {
		Reset()
		return
	}

	if nmiPending {
		nmiPending = false
		irqPending = false
		irqEnable1 = false
		call(0x0066)
		return
	}

	if !irqPending {
		return
	}
	irqPending = false
	if !irqEnable1 {
		return
	}
	irqEnable1 = false
	irqEnable2 = false
	Halted = false
	switch irqMode {
	case IrqMode0:
		panic("irq mode 0 not implemented")
	case IrqMode1:
		call(0x38)
	case IrqMode2:
		low := irqData & 0xfe
		ind := word(irqPage, low)
		loc := ref(ind).Rd16()
		call(loc)
	}
}
