package z80

// A Reg is the value of an 8-bit register.
type Reg uint8

// Rd implements the ByteRef interface.
func (r *Reg) Rd() uint8 { return uint8(*r) }

// Wr implements the ByteRef interface.
func (r *Reg) Wr(v uint8) { *r = Reg(v) }

// A Reg16 is the value of a 16-bit register.
type Reg16 uint16

// Rd16 implements the WordRef interface.
func (r *Reg16) Rd16() uint16 { return uint16(*r) }

// Wr16 implements the WordRef interface.
func (r *Reg16) Wr16(w uint16) { *r = Reg16(w) }

// A RegPair references a pair of 8-bit registers that
// can act as a single 16-bit register.
type RegPair struct {
	hi *Reg
	lo *Reg
}

// Rd16 implements the WordRef interface.
func (rp RegPair) Rd16() uint16 {
	return word(rp.hi.Rd(), rp.lo.Rd())
}

// Wr16 implements the WordRef interface.
func (rp RegPair) Wr16(w uint16) {
	hi, lo := unword(w)
	rp.lo.Wr(lo)
	rp.hi.Wr(hi)
}

// An HlMuxer is a reference to HL, IX or IY.
type HlMuxer struct{}

var hlMuxer = HlMuxer{}

func (HlMuxer) Rd16() uint16 {
	return hlMux.Rd16()
}

func (HlMuxer) Wr16(v uint16) {
	hlMux.Wr16(v)
}

// An HlIndMuxer is a reference to the memory location pointed to by HL, IX+index or IY+index.
type HlIndMuxer struct{}

var hlIndMuxer = HlIndMuxer{}

// Rd implements the ByteRef interface
func (hlIndMuxer HlIndMuxer) Rd() uint8 {
	return hlIndMuxer.addr().Rd()
}

// Wr implements the ByteRef interface
func (hlIndMuxer HlIndMuxer) Wr(v uint8) {
	hlIndMuxer.addr().Wr(v)
}

// addr is a helper that returns the relevant memory reference
func (hlIndMuxer HlIndMuxer) addr() Addr {
	loc := hlMux.Rd16()
	if indexMode != IndexModeNone {
		if !gotIndex {
			index = imm8()
			gotIndex = true
		}
		loc += uint16(index)
	}
	return ref(loc)
}

// reg selects an 8-bit register
var reg = [8]ByteRef{
	b, c, d, e, h, l, hlIndMuxer, a,
}

// qq_reg selects a 16-bit register from bc,de,hl,af
var qq_reg = [4]WordRef{bc, de, hlMuxer, af}

// dd_reg selects a 16-bit register from bc,de,hl,sp
var dd_reg = [4]WordRef{bc, de, hlMuxer, sp}
