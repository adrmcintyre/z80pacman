package main

var bitTable = newBitTable()

// CB prefix
func newBitTable() *Table {
	t := newTable()

	for rrr := range 8 {
		ea := reg(rrr)
		for bit := range 8 {
			mask := uint8(1) << bit
			t.def(
				0b01_000_000|bit<<3|rrr,
				func() {
					v := ea.Rd()
					flagZ.put((v & mask) == 0)
					flagH.set()
					flagN.reset()
				},
				"bit %b,%r", bit, rrr)

			t.def(
				0b11_000_000|bit<<3|rrr, func() {
					v := ea.Rd()
					ea.Wr(v | mask)
				},
				"set %b,%r", bit, rrr)

			t.def(
				0b10_000_000|bit<<3|rrr, func() {
					v := ea.Rd()
					ea.Wr(v & ^mask)
				},
				"res %b,%r", bit, rrr)
		}

		t.def(
			0b00_000_000|rrr,
			func() { setExtRotFlags(rlc(ea)) },
			"rlc %r", rrr)
		t.def(
			0b00_001_000|rrr,
			func() { setExtRotFlags(rrc(ea)) },
			"rrc %r", rrr)
		t.def(
			0b00_010_000|rrr,
			func() { setExtRotFlags(rl(ea, flagC.get())) },
			"rl %r", rrr)
		t.def(
			0b00_011_000|rrr,
			func() { setExtRotFlags(rr(ea)) },
			"rr %r", rrr)

		t.def(
			0b00_100_000|rrr,
			func() { setExtRotFlags(rl(ea, false)) },
			"sla %r", rrr)
		t.def(
			0b00_101_000|rrr,
			func() { setExtRotFlags(sra(ea)) },
			"sra %r", rrr)
		t.def(
			0b00_110_000|rrr,
			func() { setExtRotFlags(rl(ea, true)) },
			"sll %r", rrr)
		t.def(
			0b00_111_000|rrr,
			func() { setExtRotFlags(srl(ea)) },
			"srl %r", rrr)
	}

	return t
}
