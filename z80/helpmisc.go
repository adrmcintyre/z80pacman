package z80

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
