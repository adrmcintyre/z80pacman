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
				0b11_000_000|bit<<3|rrr, func() {
					v := ea.Rd()
					ea.Wr(v & ^mask)
				},
				"res %b,%r", bit, rrr)
		}

		t.def(
			0b00_000_000|rrr<<3,
			func() { setExtRotFlags(rlc(ea)) },
			"rlc %r", rrr)
		t.def(
			0b00_000_001|rrr<<3,
			func() { setExtRotFlags(rrc(ea)) },
			"rrc %r", rrr)
		t.def(
			0b00_000_010|rrr<<3,
			func() { setExtRotFlags(rl(ea, flagC.get())) },
			"rl %r", rrr)
		t.def(
			0b00_000_011|rrr<<3,
			func() { setExtRotFlags(rr(ea)) },
			"rr %r", rrr)

		t.def(
			0b00_000_100|rrr<<3,
			func() { setExtRotFlags(rl(ea, false)) },
			"sla %r", rrr)
		t.def(
			0b00_000_101|rrr<<3,
			func() { setExtRotFlags(sra(ea)) },
			"sra %r", rrr)
		t.def(
			0b00_000_110|rrr<<3,
			func() { setExtRotFlags(rl(ea, true)) },
			"sll %r", rrr)
		t.def(
			0b00_000_111|rrr<<3,
			func() { setExtRotFlags(srl(ea)) },
			"srl %r", rrr)
	}

	return t
}
