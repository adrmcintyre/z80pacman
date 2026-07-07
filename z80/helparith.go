package z80

// adc computes x+y+ci and returns the result.
// Affects C,S,Z,PV,N,H flags.
func adc(x uint8, y uint8, ci bool) uint8 {
	res := x + y
	if ci {
		res += 1
		flagC.put(x >= ^y)
	} else {
		flagC.put(x > ^y)
	}
	carries := res ^ x ^ y
	c4 := (carries & (1 << 4)) != 0
	c7 := (carries & (1 << 7)) != 0
	flagS.put((res & (1 << 7)) != 0)
	flagZ.put(res == 0)
	flagH.put(c4)
	flagPV.put(c7 != flagC.get())
	flagN.reset()
	return res
}

// sbc computes x-y-ci and returns the result.
// Affects C,S,Z,PV,N,H flags.
func sbc(x uint8, y uint8, ci bool) uint8 {
	res := adc(x, ^y, !ci)
	flagC.invert()
	flagN.set()
	return res
}

// inc increments the byte at ea.
// Affects S,Z,PV,N,H flags.
func inc(ea ByteRef) {
	data := ea.Rd()
	res := data + 1
	flagS.put((res & (1 << 7)) != 0)
	flagZ.put(res == 0)
	flagPV.put(res == 0)
	flagH.put((data & 0x0f) == 0x0f)
	flagN.reset()
	ea.Wr(res)
}

// dec decrements the byte at ea.
// Affects S,Z,PV,N,H flags.
func dec(ea ByteRef) {
	data := ea.Rd()
	res := data - 1
	flagS.put((res & (1 << 7)) != 0)
	flagZ.put(res == 0)
	flagPV.put(res == 0xff)
	flagH.put((data & 0x0f) == 0x0f)
	flagN.set()
	ea.Wr(res)
}

// add16 computes w1+w2 and returns the result.
// Affects C,N flags.
func add16(w1 uint16, w2 uint16) uint16 {
	res := w1 + w2
	flagC.put(res < w1)
	flagN.reset()
	return res
}
