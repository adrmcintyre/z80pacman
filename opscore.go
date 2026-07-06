package main

// This file defines the core (non-prefixed) z80 opcodes.

var coreTable = newCoreTable()

func newCoreTable() *opTable {
	t := newTable()

	for rrr := range 8 {
		dst := reg(rrr)
		for sss := range 8 {
			src := reg(sss)
			t.def(
				0b01_000_000|(rrr<<3)|sss,
				func() { dst.Wr(src.Rd()) },
				"ld %r,%r", rrr, sss)

		}
	}
	for rrr := range 8 {
		dst := reg(rrr)
		ld_r_n := func() { dst.Wr(imm8()) }
		if rrr == 6 {
			// Special case to cover "ld (ix+#d),n" / "ld (iy+#d),n"
			// we need to ensure the displacement <d> is read *before* the value <n>
			ld_r_n = func() {
				ea := hlIndMuxer.addr()
				ea.Wr(imm8())
			}
		}
		t.def(0b00_000_110|(rrr<<3), ld_r_n, "ld %r,%n", rrr)
	}
	// overwrite 0x76 with hlt (would be ld (hl),(hl))
	t.def(0b01_110_110, func() { halted = true }, "hlt")

	t.def(0b00_000_010, func() { mem(bc.Rd16()).Wr(a.Rd()) }, "ld (bc),a")
	t.def(0b00_001_010, func() { a.Wr(mem(bc.Rd16()).Rd()) }, "ld a,(bc)")
	t.def(0b00_010_010, func() { mem(de.Rd16()).Wr(a.Rd()) }, "ld (de),a")
	t.def(0b00_011_010, func() { a.Wr(mem(de.Rd16()).Rd()) }, "ld a,(de)")

	t.def(0b00_100_010, func() { mem(imm16()).Wr16(hlMux.Rd16()) }, "ld (%N),%h")
	t.def(0b00_101_010, func() { hlMux.Wr16(mem(imm16()).Rd16()) }, "ld %h,(%N)")
	t.def(0b00_110_010, func() { mem(imm16()).Wr(a.Rd()) }, "ld (%N),a")
	t.def(0b00_111_010, func() { a.Wr(mem(imm16()).Rd()) }, "ld a,(%N)")

	for i := range 4 {
		dst := dd(i)
		t.def(
			0b00_000_001|(i<<4),
			func() { dst.Wr16(imm16()) },
			"ld %d,%N", i)
	}
	t.def(
		0b11_111_001,
		func() { sp.Wr16(hlMux.Rd16()) },
		"ld sp,%h")

	for i := range 4 {
		// push qq
		ea := qq(i)
		t.def(
			0b11_000_101|i<<4,
			func() { push16(ea.Rd16()) },
			"push %q", i)

		// pop qq
		t.def(
			0b11_000_001|i<<4,
			func() { ea.Wr16(pop16()) },
			"pop %q", i)
	}

	t.def(
		0b00_001_000, func() {
			tmp := af.Rd16()
			af.Wr16(af2.Rd16())
			af2.Wr16(tmp)
		},
		"ex af,af'")

	t.def(
		0b11_011_001, func() {
			tmp := bc.Rd16()
			bc.Wr16(bc2.Rd16())
			bc2.Wr16(tmp)

			tmp = de.Rd16()
			de.Wr16(de2.Rd16())
			de2.Wr16(tmp)

			// NOTE - always HL even if prefixed
			tmp = hl.Rd16()
			hl.Wr16(hl2.Rd16())
			hl2.Wr16(tmp)
		},
		"exx")

	t.def(
		0b11_100_011, func() {
			addr := mem(sp.Rd16())
			tmp := hlMux.Rd16()
			data := addr.Rd16()
			addr.Wr16(tmp)
			hlMux.Wr16(data)
		},
		"ex (sp),%h")

	t.def(
		0b11_101_011, func() {
			tmp := de.Rd16()
			// NOTE - always HL even if prefixed
			de.Wr16(hl.Rd16())
			hl.Wr16(tmp)
		},
		"ex de,hl")

	for i := range 9 {
		var (
			src    ByteRef
			opcode int
			arg    string
		)
		if i < 8 {
			// <op> a,r
			src = reg(i)
			opcode = 0b10_000_000 | i
			arg = "%r"
		} else {
			// <op> a,imm
			src = ByteRef(imm)
			opcode = 0b11_000_110
			arg = "%n"
		}

		t.def(opcode|0b000<<3, func() { a.Wr(adc(a.Rd(), src.Rd(), false)) }, "add a,"+arg, i)
		t.def(opcode|0b001<<3, func() { a.Wr(adc(a.Rd(), src.Rd(), flagC.get())) }, "adc a,"+arg, i)
		t.def(opcode|0b010<<3, func() { a.Wr(sbc(a.Rd(), src.Rd(), false)) }, "sub "+arg, i)
		t.def(opcode|0b011<<3, func() { a.Wr(sbc(a.Rd(), src.Rd(), flagC.get())) }, "sbc "+arg, i)

		t.def(opcode|0b100<<3, func() { a.Wr(a.Rd() & src.Rd()); setLogicFlags(); flagH.set() }, "and "+arg, i)
		t.def(opcode|0b101<<3, func() { a.Wr(a.Rd() ^ src.Rd()); setLogicFlags(); flagH.reset() }, "xor "+arg, i)
		t.def(opcode|0b110<<3, func() { a.Wr(a.Rd() | src.Rd()); setLogicFlags(); flagH.reset() }, "or "+arg, i)

		t.def(opcode|0b111<<3, func() { _ = sbc(a.Rd(), src.Rd(), false) }, "cp "+arg, i)
	}

	for rrr := range 8 {
		ea := reg(rrr)
		t.def(0b00_000_100|rrr<<3, func() { inc(ea) }, "inc %r", rrr)
		t.def(0b00_000_101|rrr<<3, func() { dec(ea) }, "dec %r", rrr)
	}

	t.def(
		0b00_100_111, func() {
			value := a.Rd()
			hIn := flagH.get()
			cIn := flagC.get()
			if flagN.get() {
				if hIn || ((value & 0x0f) > 0x09) {
					value -= 0x06
				}
				if cIn || ((value & 0xf0) > 0x90) {
					value -= 0x60
					flagC.set()
				}
				if hIn {
					flagH.put((value & 0x0f) < 6)
				}
			} else {
				if hIn || ((value & 0xf) > 0x09) {
					value += 0x06
				}
				if cIn || ((value & 0xf0) > 0x90) {
					value += 0x60
					flagC.set()
				}
				flagH.put((value & 0x0f) >= 0x0a)
			}
			a.Wr(value)
			flagS.put((value & (1 << 7)) != 0)
			flagZ.put(value == 0)
			setParity(value)
		},
		"daa")

	t.def(
		0b00_101_111, func() {
			a.Wr(^a.Rd())
			flagH.set()
			flagN.set()
		},
		"cpl")

	t.def(0b00_110_111, func() { flagC.set() }, "scf")
	t.def(0b00_111_111, func() { flagC.invert() }, "ccf")

	t.def(0b00_000_000, func() {}, "nop")

	t.def(0b11_110_011, func() { irqEnable1 = false }, "di")
	t.def(0b11_111_011, func() {
		irqEnable1 = true
		irqEnable2 = true
		//TODO - note, any pending interrupt should not be
		//accepted until *after* the instruction following this
	}, "ei")

	for i := range 4 {
		ea := dd(i)
		t.def(0b00_001_001|i<<4, func() { hlMux.Wr16(add16(hlMux.Rd16(), ea.Rd16())) }, "add %h,%d", i)
		t.def(0b00_000_011|i<<4, func() { ea.Wr16(ea.Rd16() + 1) }, "inc %d", i)
		t.def(0b00_001_011|i<<4, func() { ea.Wr16(ea.Rd16() - 1) }, "dec %d", i)
	}

	t.def(0b00_000_111, func() { setRotFlags(rlc(a)) }, "rlca")
	t.def(0b00_001_111, func() { setRotFlags(rrc(a)) }, "rrca")
	t.def(0b00_010_111, func() { setRotFlags(rl(a, flagC.get())) }, "rla")
	t.def(0b00_011_111, func() { setRotFlags(rr(a)) }, "rra")

	t.def(0b11_000_011, func() { jmp(imm16()) }, "jp %N")

	t.def(0b11_001_101, func() { call(imm16()) }, "call %N")
	t.def(0b11_001_001, ret, "ret")

	for cond := range 4 {
		for value := range 2 {
			t.def(
				0b11_000_010|cond<<4|value<<3, func() {
					nn := imm16()
					if conditionIs(cond, value) {
						jmp(nn)
					}
				},
				"jp %c,%N", cond<<1|value)
			t.def(
				0b11_000_100|cond<<4|value<<3, func() {
					nn := imm16()
					if conditionIs(cond, value) {
						call(nn)
					}
				},
				"call %c,%N", cond<<1|value)
			t.def(
				0b11_000_000|cond<<4|value<<3, func() {
					if conditionIs(cond, value) {
						ret()
					}
				},
				"ret %c", cond<<1|value)
		}
	}

	t.def(
		0b00_011_000, func() {
			jr(imm8())
		},
		"jr %e")

	for cond := range 2 {
		for value := range 2 {
			t.def(
				0b00_100_000|cond<<4|value<<3, func() {
					offset := imm8()
					if conditionIs(cond, value) {
						jr(offset)
					}
				},
				"jr %c,%e", cond<<1|value)
		}
	}

	t.def(
		0b11_101_001, func() {
			jmp(hlMux.Rd16())
		},
		"jp (%h)")

	t.def(
		0b00_010_000, func() {
			offset := imm8()
			v := b.Rd()
			v -= 1
			b.Wr(v)
			if v != 0 {
				jr(offset)
			}
		}, "djnz %e")

	/// rst p
	for i := range 8 {
		loc := uint16(i * 8)
		t.def(
			0b11_000_111|i<<3, func() {
				call(loc)
			},
			"rst %p", i)
	}

	// out (n),A
	t.def(
		0b11_010_011, func() {
			// The operand n is placed on the bottom half (A0 through A7) of the
			// address bus to select the I/O device at one of 256 possible ports.
			// The contents of the Accumulator (Register A) also appear on the
			// top half (A8 through A15) of the address bus at this time. Then
			// the byte contained in the Accumulator is placed on the data bus
			// and written to the selected peripheral device.
			//
			// If the Accumulator contains 23h, then upon the execution of an
			// OUT (01h) instruction, byte 23h is written to the peripheral device
			// mapped to I/O port address 01h.
			n := imm8()
			v := a.Rd()
			addressBus.Store(uint32(word(v, n)))
			dataBus.Store(uint32(v))
			ioWrite()

			// Pacman only uses this in two cases to set up an interrupt vector.
			//
			//	After program start:
			//
			//		ld  i, 0x3f
			//		...
			//		ld  a, 0xfa
			// 		out (0), a  // sets 0x3ffa - contains 0x3000 (CheckROM)
			//		...
			//		halt // wait for an interrupt
			//
			// After start-up tests have passed:
			//
			//		xor a
			//		sub 4
			//		out (0),a   // sets 0x3ffc - contains 0x008d (VBLANK every 1/60.61s)
			//		...
			//		// enter main program
		},
		"out (%n),a")

	return t
}
