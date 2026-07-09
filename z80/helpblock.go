package z80

// blockTx performs the following operation:
// (DE) <- (HL)     ; copy source to dest
// DE <- DE + delta ; inc/dec dest pointer
// HL <- HL + delta ; inc/dec source pointer
// BC <- BC – 1     ; decrement counter
// Affects H,PV,N flags.
func blockTx(delta uint16, repeat bool) {
	src := hl.Rd16()
	dst := de.Rd16()
	count := bc.Rd16()

	v := ref(src).Rd()
	ref(dst).Wr(v)
	dst += delta
	src += delta
	count -= 1

	hl.Wr16(src)
	de.Wr16(dst)
	bc.Wr16(count)

	flagH.reset()
	flagN.reset()
	flagPV.put(count != 0)

	if repeat && count != 0 {
		jmp(pc.Rd16() - 2)
	}
}

// blockCp performs the following operation:
// A - (HL)         ; compare A with source
// HL <- HL + delta ; inc/dec source pointer
// BC <- BC – 1     ; decrement counter
// Affects S,Z,H,PV,N flags.
func blockCp(delta uint16, repeat bool) {
	src := hl.Rd16()
	count := bc.Rd16()

	x := a.Rd()
	y := ref(src).Rd()
	src += delta
	count -= 1

	hl.Wr16(src)
	bc.Wr16(count)

	diff := x - y
	c4 := (diff^x^y)&(1<<4) != 0

	flagS.put(diff&(1<<7) != 0)
	flagZ.put(diff == 0)
	flagH.put(c4)
	flagPV.put(count != 0)
	flagN.set()

	if repeat && !(count == 0 || diff == 0) {
		jmp(pc.Rd16() - 2)
	}
}
