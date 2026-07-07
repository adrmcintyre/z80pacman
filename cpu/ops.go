package cpu

// This file implements helpers for various z80 opcodes
// and flag manipulation.

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

// setLogicFlags sets common flags affected by the logic operations and, or, xor.
// Affects C,S,Z,PV,N flags.
func setLogicFlags() {
	v := a.Rd()
	flagS.put((v & (1 << 7)) != 0)
	flagZ.put(v == 0)
	setParity(v)
	flagN.reset()
	flagC.reset()
}

// setParity sets the PV flag according to the parity of the argument.
func setParity(v uint8) {
	v ^= v >> 1
	v ^= v >> 2
	v ^= v >> 4
	flagPV.put((v & 1) == 0)
}

// add16 computes w1+w2 and returns the result.
// Affects C,N flags.
func add16(w1 uint16, w2 uint16) uint16 {
	res := w1 + w2
	flagC.put(res < w1)
	flagN.reset()
	return res
}

// push16 pushes a word to the stack.
func push16(w uint16) {
	loc := sp.Rd16() - 2
	sp.Wr16(loc)
	ref(loc).Wr16(w)
}

// pop16 returns a word popped from the stack.
func pop16() uint16 {
	loc := sp.Rd16()
	sp.Wr16(loc + 2)
	return ref(loc).Rd16()
}

// jmp sets the program counter to nn.
func jmp(nn uint16) {
	pc.Wr16(nn)
}

// jr adjusts the program counter by the given 8-bit relative offset.
func jr(offset uint8) {
	loc := pc.Rd16() + uint16(offset)
	if offset >= 0x80 {
		loc -= 0x100
	}
	jmp(loc)
}

// call pushes the program counter and sets the program counter to nn.
func call(nn uint16) {
	push16(pc.Rd16())
	jmp(nn)
}

// ret pops a word (return address) from the stack and jumps to it.
func ret() {
	jmp(pop16())
}

// rlc rotates the contents of ea left 1 bit position.
// Bit 7 is copied to the C flag and also to bit 0.
// Affects C,N,H flags.
func rlc(ea ByteRef) uint8 {
	v := ea.Rd()
	v7 := v >> 7
	res := v<<1 | v7
	ea.Wr(res)
	setRotFlags(v7 != 0)
	return res
}

// rrc rotates the contents of ea right 1 bit position.
// Bit 0 is copied to the C flag and also to bit 7.
// Affects C,N,H flags.
func rrc(ea ByteRef) uint8 {
	v := ea.Rd()
	v0 := v & 1
	res := v>>1 | v0<<7
	ea.Wr(res)
	setRotFlags(v0 != 0)
	return res
}

// rl rotates the contents of ea left 1 bit position.
// Bit 7 is copied to the C flag, and ci is copied to bit 0.
// Affects C,N,H flags.
func rl(ea ByteRef, ci bool) uint8 {
	v := ea.Rd()
	v7 := v >> 7
	res := v << 1
	if ci {
		res |= 1
	}
	ea.Wr(res)
	setRotFlags(v7 != 0)
	return res
}

// rr rotates the contents of ea right 1 bit position through the C flag.
// Bit 0 is copied to the C flag, and ci is copied to bit 7.
// Affects C,N,H flags.
func rr(ea ByteRef, ci bool) uint8 {
	v := ea.Rd()
	v0 := v & 1
	res := v >> 1
	if ci {
		res |= 1 << 7
	}
	ea.Wr(res)
	setRotFlags(v0 != 0)
	return res
}

func sla(ea ByteRef) uint8 {
	v := ea.Rd()
	v7 := v >> 7
	res := v << 1
	ea.Wr(res)
	setRotFlags(v7 != 0)
	return res
}

// sra shifts the contents of ea right 1 bit position. Bit 0 is copied to
// the C flag. The contents of bit 7 remain unchanged.
// Affects C,N,H flags.
func sra(ea ByteRef) uint8 {
	v := ea.Rd()
	v0 := v & 1
	res := (v >> 1) | (v & (1 << 7))
	ea.Wr(res)
	setRotFlags(v0 != 0)
	return res
}

// srl shifts the contents of ea right 1 bit position.
// Bit 0 is copied to the C flag.
// Affects C,N,H flags.
func srl(ea ByteRef) uint8 {
	v := ea.Rd()
	v0 := v & 1
	res := v >> 1
	ea.Wr(res)
	setRotFlags(v0 != 0)
	return res
}

// setRotFlags sets common flags used by the shift and rotate instructions.
// Sets the C flag is set according to co; clears the H and N flags.
func setRotFlags(co bool) {
	flagC.put(co)
	flagH.reset()
	flagN.reset()
}

// setExtRotFlags sets additional flags used by extended shift and rotate instructions,
// according to the result of the shift supplied by v.
// Affects S,Z,PV flags.
func setExtRotFlags(v uint8) {
	flagS.put(v&(1<<7) != 0)
	flagZ.put(v == 0)
	setParity(v)
}
