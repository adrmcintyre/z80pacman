package main

// Addr is a reference to a memory location.
type Addr uint16

// Rd implements the ByteRef interface.
func (addr Addr) Rd() uint8 {
	return memRead(uint16(addr))
}

// Wr implements the ByteRef interface.
func (addr Addr) Wr(v uint8) {
	memWrite(uint16(addr), v)
}

// Rd16 implements the WordRef interface.
func (addr Addr) Rd16() uint16 {
	lo := addr.Rd()
	hi := (addr + 1).Rd()
	return word(hi, lo)
}

// Wr16 implements the WordRef interface.
func (addr Addr) Wr16(w uint16) {
	hi, lo := unword(w)
	addr.Wr(lo)
	(addr + 1).Wr(hi)
}

// mem returns a reference to the specified memory address.
func mem(addr uint16) Addr {
	return Addr(addr)
}
