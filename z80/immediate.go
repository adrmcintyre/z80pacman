package z80

// An Imm is a reference to the instruction stream
type Imm struct{}

var imm = Imm{}

// Implements the ByteRef interface
func (Imm) Rd() uint8 {
	return imm8()
}

// Implements the ByteRef interface
func (Imm) Wr(uint8) {
	panic("illegal write to immediate")
}

// Implements the WordRef interface
func (Imm) Rd16() uint16 {
	return imm16()
}

// Implements the WordRef interface
func (Imm) Wr16(uint16) {
	panic("illegal write to immediate")
}

// imm8 reads a byte from the instruction stream
func imm8() uint8 {
	loc := pc.Rd16()
	n := ref(loc).Rd()
	DebugTrace = append(DebugTrace, n)
	pc.Wr16(loc + 1)
	return n
}

// imm16 reads a word from the instruction stream (lo,hi)
func imm16() uint16 {
	nlo := imm8()
	nhi := imm8()
	return word(nhi, nlo)
}
