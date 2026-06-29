package main

type flagBit int

const (
	flagC  = flagBit(1 << 0) // carry
	flagN  = flagBit(1 << 1) // add/subtract
	flagPV = flagBit(1 << 2) // overflow/parity - see page 67
	_      = flagBit(1 << 3) // (X) unused
	flagH  = flagBit(1 << 4) // half carry - see page 68
	_      = flagBit(1 << 5) // (Y) unused
	flagZ  = flagBit(1 << 6) // zero - see page 68
	flagS  = flagBit(1 << 7) // sign - see page 69
)

func (fl flagBit) get() bool {
	return f.Rd()&uint8(fl) != 0
}

func (fl flagBit) set() {
	f.Wr(f.Rd() | uint8(fl))
}

func (fl flagBit) reset() {
	f.Wr(f.Rd() & ^uint8(fl))
}

func (fl flagBit) invert() {
	f.Wr(f.Rd() ^ uint8(fl))
}

func (fl flagBit) put(v bool) {
	if v {
		fl.set()
	} else {
		fl.reset()
	}
}

var flagTable = [4]flagBit{
	flagZ,
	flagC,
	flagPV,
	flagS,
}

func conditionIs(cc int, v int) bool {
	return flagTable[cc].get() == (v != 0)
}
