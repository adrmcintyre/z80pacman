package z80

// This file defines the extended ("ED"-prefixed) z80 opcodes.

var miscTable = newMiscTable()

func newMiscTable() *opTable {
	t := newTable()

	for rrr := range 8 {
		if rrr == 6 {
			// no (hl) mode for these
			continue
		}
		r := reg[rrr]
		_ = r
		t.def(0b01_000_000|rrr<<3, func() { unimplemented() }, "in %r,(c)", rrr)
		t.def(0b01_000_001|rrr<<3, func() { unimplemented() }, "out (c),%r", rrr)
	}

	for dd := range 4 {
		ea := dd_reg[dd]
		t.def(
			0b01_000_010|dd<<4, func() {
				x := hlMux.Rd16()
				y := ea.Rd16()

				xHi, xLo := unword(x)
				yHi, yLo := unword(y)

				resLo := sbc(xLo, yLo, flagC.get())
				resHi := sbc(xHi, yHi, flagC.get())
				flagZ.put(resLo == 0 && resHi == 0)

				hlMux.Wr16(word(resHi, resLo))
			},
			"sbc %h,%d", dd)

		t.def(
			0b01_001_010|dd<<4, func() {
				x := hlMux.Rd16()
				y := ea.Rd16()

				xHi, xLo := unword(x)
				yHi, yLo := unword(y)

				resLo := adc(xLo, yLo, flagC.get())
				resHi := adc(xHi, yHi, flagC.get())
				flagZ.put(resLo == 0 && resHi == 0)

				hlMux.Wr16(word(resHi, resLo))
			},
			"adc %h,%d", dd)

		t.def(
			0b01_000_011|dd<<4, func() {
				loc := imm16()
				w := ea.Rd16()
				ref(loc).Wr16(w)
			},
			"ld (%N),%d", dd)

		t.def(
			0b01_001_011|dd<<4, func() {
				loc := imm16()
				w := ref(loc).Rd16()
				ea.Wr16(w)
			},
			"ld %d,(%N)", dd)
	}

	t.def(
		0b01_000_100, func() {
			a.Wr(sbc(0, a.Rd(), false))
		},
		"neg")

	t.def(
		0b01_000_101, func() {
			irqEnable1 = irqEnable2
			ret()
		},
		"retn")

	t.def(0b01_001_101, ret, "reti")

	// im n (note: encoded as 0, 2, 3)
	for mode := range 3 {
		encodedMode := mode
		if mode > 0 {
			encodedMode += 1
		}
		t.def(
			0b01_000_110|encodedMode<<3, func() {
				irqMode = IrqMode(mode)
			},
			"im %m", mode)
	}

	t.def(
		0b01_000_111, func() {
			irqPage = a.Rd()
		},
		"ld i,a")

	t.def(
		0b01_010_111, func() {
			a.Wr(irqPage)
			flagPV.put(irqEnable2)
		},
		"ld a,i")

	t.def(0b01_001_111, unimplemented, "ld r,a")
	t.def(0b01_011_111, unimplemented, "ld a,r")
	t.def(0b01_100_111, rrd, "rrd")
	t.def(0b01_101_111, rld, "rld")

	ldop := []string{"ldi", "ldd", "ldir", "lddr"}
	cpop := []string{"cpi", "cpd", "cpir", "cpdr"}
	inop := []string{"ini", "ind", "inir", "indr"}
	outop := []string{"outi", "outd", "otir", "otdr"}
	dirDelta := []uint16{1, 0xffff}

	for repeatBit := range 2 {
		repeat := repeatBit == 1
		for dirBit := range 2 {
			delta := dirDelta[dirBit]
			index := repeatBit*2 + dirBit

			//0b10_1rd_000 // ldi ldd ldir lddr
			t.def(
				0b10_100_000|repeatBit<<4|dirBit<<3, func() {
					blockTx(delta, repeat)
				},
				ldop[index])

			//0b10_1rd_001 // cpi cpd cpir cpdr
			t.def(
				0b10_100_001|repeatBit<<4|dirBit<<3, func() {
					blockCp(delta, repeat)
				},
				cpop[index])

			//0b10_1rd_010 // ini ind inir indr
			t.def(
				0b10_100_010|repeatBit<<4|dirBit<<3,
				unimplemented,
				inop[index])

			//0b10_1rd_011 // outi outd otir otdr
			t.def(
				0b10_100_011|repeatBit<<4|dirBit<<3,
				unimplemented,
				outop[index])
		}
	}

	return t
}
