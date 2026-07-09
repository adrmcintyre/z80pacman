package z80

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

// rld treats the lower nibble of A and the contents of the byte pointed
// to by HL as a 12 bit word, which it then rotates left by 4 bits.
//
// (HL)[7..4] <- (HL)[3..0] <- A[3..0] <- (HL)[7..4]
// Affects S,Z,H,PV,N flags.
func rld() {
	loc := ref(hl.Rd16())
	memOld := loc.Rd()
	aOld := a.Rd()
	memNew := (memOld << 4) | (aOld & 0x0f)
	aNew := (aOld & 0xf0) | (memNew >> 4)
	a.Wr(aNew)
	loc.Wr(memNew)

	flagS.put((aNew & 1 << 7) != 0)
	flagZ.put(aNew == 0)
	flagH.reset()
	flagPV.put(parity(aNew))
	flagN.reset()
}

// rrd treats the lower nibble of A and the contents of the byte pointed
// to by HL as a 12 bit word, which it then rotates right by 4 bits.
//
// (HL)[3..0] <- (HL)[7..4] <- A[3..0] <- (HL)[3..0]
// Affects S,Z,H,PV,N flags.
func rrd() {
	loc := ref(hl.Rd16())
	memOld := loc.Rd()
	aOld := a.Rd()
	memNew := (aOld << 4) | (memOld >> 4)
	aNew := (aOld & 0xf0) | (memOld & 0x0f)
	a.Wr(aNew)
	loc.Wr(memNew)

	flagS.put((aNew & 1 << 7) != 0)
	flagZ.put(aNew == 0)
	flagH.reset()
	flagPV.put(parity(aNew))
	flagN.reset()
}
