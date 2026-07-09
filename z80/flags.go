package z80

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

// get returns true if the corresponding flag bit in the F register is set.
func (fl flagBit) get() bool {
	return f.Rd()&uint8(fl) != 0
}

// set sets the the corresponding flag bit in the F register.
func (fl flagBit) set() {
	f.Wr(f.Rd() | uint8(fl))
}

// reset clears the corresponding flag bit in the F register.
func (fl flagBit) reset() {
	f.Wr(f.Rd() & ^uint8(fl))
}

// invert negates the corresponding flag bit in the F register.
func (fl flagBit) invert() {
	f.Wr(f.Rd() ^ uint8(fl))
}

// put replaces the corresponding flag bit in the F register with v.
func (fl flagBit) put(v bool) {
	if v {
		fl.set()
	} else {
		fl.reset()
	}
}

// flagTable contains a list of flagBits to aid decoding condition codes.
var flagTable = [4]flagBit{
	flagZ,
	flagC,
	flagPV,
	flagS,
}

// conditionIs returns true if the flag corresponding to the given condition
// code's bit pattern in cc has the value in v.
func conditionIs(cc int, v int) bool {
	return flagTable[cc].get() == (v != 0)
}

// parity returns true if v has an even number of 1 bits.
func parity(v uint8) bool {
	v ^= v >> 1
	v ^= v >> 2
	v ^= v >> 4
	return (v & 1) == 0
}

// setLogicFlags sets common flags affected by the logic operations and, or, xor.
// Affects C,S,Z,PV,N flags.
func setLogicFlags() {
	v := a.Rd()
	flagS.put((v & (1 << 7)) != 0)
	flagZ.put(v == 0)
	flagPV.put(parity(v))
	flagN.reset()
	flagC.reset()
}

// setRotFlags sets common flags used by the shift and rotate instructions.
// Sets the C flag according to co; clears the H and N flags.
func setRotFlags(co bool) {
	flagC.put(co)
	flagH.reset()
	flagN.reset()
}

// setExtRotFlags sets additional flags used by extended shift and rotate
// instructions according to the result of the shift supplied in v.
// Affects S,Z,PV flags.
func setExtRotFlags(v uint8) {
	flagS.put(v&(1<<7) != 0)
	flagZ.put(v == 0)
	flagPV.put(parity(v))
}
