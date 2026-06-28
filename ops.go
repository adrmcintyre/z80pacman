package main

func illegal() {
	panic("illegal instruction")
}

func TODO() {
	panic("TODO")
}

func NOT_NEEDED() {
	panic("not needed")
}

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

func sbc(x uint8, y uint8, ci bool) uint8 {
	res := adc(x, ^y, !ci)
	flagC.invert()
	flagN.set()
	return res
}

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

func setLogicFlags(newH bool) {
	v := a.Rd()
	flagS.put((v & (1 << 7)) != 0)
	flagZ.put(v == 0)
	flagH.put(newH)
	setParity(v)
	flagN.reset()
	flagC.reset()
}

func setParity(v uint8) {
	v ^= v >> 1
	v ^= v >> 2
	v ^= v >> 4
	flagPV.put((v & 1) == 0)
}

func add16(w uint16, x uint16) uint16 {
	res := w + x
	flagC.put(res < w)
	flagN.reset()
	return res
}

func push16(w uint16) {
	loc := sp.Rd16() - 2
	sp.Wr16(loc)
	mem(loc).Wr16(w)
}

func pop16() uint16 {
	loc := sp.Rd16()
	sp.Wr16(loc + 2)
	return mem(loc).Rd16()
}

func jmp(nn uint16) {
	pc.Wr16(nn)
}

func jr(offset uint8) {
	loc := pc.Rd16() + uint16(offset)
	if offset >= 0x80 {
		loc -= 0x100
	}
	jmp(loc)
}

func call(nn uint16) {
	push16(pc.Rd16())
	jmp(nn)
}

func ret() {
	jmp(pop16())
}

func rlc(ea ByteRef) (uint8, bool) {
	v := ea.Rd()
	v7 := v >> 7
	res, co := v<<1|v7, v7 != 0
	ea.Wr(res)
	return res, co
}

func rrc(ea ByteRef) (uint8, bool) {
	v := ea.Rd()
	v0 := v & 1
	res, co := v>>1|v0<<7, v0 != 0
	ea.Wr(res)
	return res, co
}

func rl(ea ByteRef, ci bool) (uint8, bool) {
	v := ea.Rd()
	v7 := v >> 7
	res, co := v<<1, v7 != 0
	if ci {
		res |= 1
	}
	ea.Wr(res)
	return res, co
}

func rr(ea ByteRef) (uint8, bool) {
	v := ea.Rd()
	v0 := v & 1
	res, co := v>>1, v0 != 0
	if flagC.get() {
		res |= 1 << 7
	}
	ea.Wr(res)
	return res, co
}

func sla(ea ByteRef) (uint8, bool) {
	v := ea.Rd()
	v7 := v >> 7
	res, co := v<<1, v7 != 0
	ea.Wr(res)
	return res, co
}

func sra(ea ByteRef) (uint8, bool) {
	v := ea.Rd()
	v0 := v & 1
	res, co := (v>>1)|(v&(1<<7)), v0 != 0
	ea.Wr(res)
	return res, co
}

func srl(ea ByteRef) (uint8, bool) {
	v := ea.Rd()
	v0 := v & 1
	res, co := v>>1, v0 != 0
	ea.Wr(res)
	return res, co
}

func setRotFlags(_ uint8, co bool) {
	flagC.put(co)
	flagH.reset()
	flagN.reset()
}

func setExtRotFlags(v uint8, co bool) {
	flagS.put(v&(1<<7) != 0)
	flagZ.put(v == 0)
	setParity(v)
	flagC.put(co)
	flagH.reset()
	flagN.reset()
}
