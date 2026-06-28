package main

type flag int

const (
	flagC  flag = iota // carry
	flagN              // add/subtract
	flagPV             // overflow/parity - see page 67
	flagX              // unused

	flagH // half carry - see page 68
	flagY // unused
	flagZ // zero - see page 68
	flagS // sign - see page 69
)

func (fl flag) get() bool {
	return f.Rd()&(1<<fl) != 0
}

func (fl flag) set() {
	f.Wr(f.Rd() | 1<<fl)
}

func (fl flag) reset() {
	f.Wr(f.Rd() & ^(1 << fl))
}

func (fl flag) invert() {
	f.Wr(f.Rd() ^ (1 << fl))
}

func (fl flag) put(v bool) {
	if v {
		fl.set()
	} else {
		fl.reset()
	}
}

var flagTable = [4]flag{
	flagZ,
	flagC,
	flagPV,
	flagS,
}

func conditionIs(cc int, v int) bool {
	return flagTable[cc].get() == (v != 0)
}
