package z80

// This file defines the extended ("ED"-prefixed) z80 opcodes.

var miscTable = newMiscTable()

func newMiscTable() *opTable {
	t := newTable()

	for r := range 8 {
		// not (HL)
		if r != 6 {
			t.def(0b01_000_000|r<<3, unimplemented, "in %r,(%n)", r)
			t.def(0b01_000_001|r<<3, unimplemented, "out (%n),%r", r)
		}
	}

	for i := range 4 {
		ea := dd(i)
		t.def(
			0b01_000_010|i<<4, func() {
				x := hlMux.Rd16()
				y := ea.Rd16()

				xHi, xLo := unword(x)
				yHi, yLo := unword(y)

				resLo := sbc(xLo, yLo, flagC.get())
				resHi := sbc(xHi, yHi, flagC.get())
				flagZ.put(resLo == 0 && resHi == 0)

				hlMux.Wr16(word(resHi, resLo))
			},
			"sbc %h,%d", i)

		t.def(
			0b01_001_010|i<<4, func() {
				x := hlMux.Rd16()
				y := ea.Rd16()

				xHi, xLo := unword(x)
				yHi, yLo := unword(y)

				resLo := adc(xLo, yLo, flagC.get())
				resHi := adc(xHi, yHi, flagC.get())
				flagZ.put(resLo == 0 && resHi == 0)

				hlMux.Wr16(word(resHi, resLo))
			},
			"adc %h,%d", i)

		t.def(
			0b01_000_011|i<<4, func() {
				loc := imm16()
				w := ea.Rd16()
				ref(loc).Wr16(w)
			},
			"ld (%N),%d", i)

		t.def(
			0b01_001_011|i<<4, func() {
				loc := imm16()
				w := ref(loc).Rd16()
				ea.Wr16(w)
			},
			"ld %d,(%N)", i)
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
	t.def(0b01_100_111, unimplemented, "rrd")
	t.def(0b01_101_111, unimplemented, "rld")

	//0b10_1rd_000 // +ldi +ldir -ldd -lddr
	//0b10_1rd_001 // -cpi +cpir -cpd -cpdr
	//0b10_1rd_010 // -ini -inir -ind -indr
	//0b10_1rd_011 // -outi -otir -outd -otdr

	t.def(0b10_100_000, func() { ldi() }, "ldi")

	t.def(
		0b10_110_000, func() {
			if ldi() {
				jmp(pc.Rd16() - 2)
			}
		},
		"ldir")

	t.def(0b10_100_001, func() { cpi() }, "cpi")

	t.def(
		0b10_110_001, func() {
			if cpi() {
				jmp(pc.Rd16() - 2)
			}
		},
		"cpir")

	return t
}

// ldi performs the following operation:
// (DE) <- (HL) ; copy source to dest
// DE <- DE + 1 ; increment dest pointer
// HL <- HL + 1 ; increment source pointer
// BC <- BC – 1 ; decrement counter
// Affects H,PV,N flags.
func ldi() bool {
	src := hl.Rd16()
	dst := de.Rd16()
	count := bc.Rd16()

	v := ref(src).Rd()
	ref(dst).Wr(v)
	dst += 1
	src += 1
	count -= 1

	hl.Wr16(src)
	de.Wr16(dst)
	bc.Wr16(count)

	(flagH | flagN).reset()
	flagPV.put(count != 0)

	return count != 0
}

// cpi performs the following operation:
// A - (HL)     ; compare A with source
// HL <- HL +1  ; increment source pointer
// BC <- BC – 1 ; decrement counter
// Affects S,Z,H,PV,N flags.
func cpi() bool {
	src := hl.Rd16()
	count := bc.Rd16()

	x := a.Rd()
	y := ref(src).Rd()
	src += 1
	count -= 1

	hl.Wr16(src)
	bc.Wr16(count)

	delta := x - y
	c4 := (delta^x^y)&(1<<4) != 0

	flagS.put(delta&(1<<7) != 0)
	flagZ.put(delta == 0)
	flagH.put(c4)
	flagPV.put(count != 0)
	flagN.set()

	return !(count == 0 || delta == 0)
}
