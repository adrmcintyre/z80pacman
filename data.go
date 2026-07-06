package main

// A ByteRef can read or write a byte of data.
type ByteRef interface {
	Rd() uint8
	Wr(v uint8)
}

// A WordRef can read or write 2 bytes of data.
type WordRef interface {
	Rd16() uint16
	Wr16(uint16)
}

// word is a helper that converts to bytes to a 16-bit word.
func word(hi, lo uint8) uint16 {
	return (uint16(hi) << 8) | uint16(lo)
}

// unword is a helper that converts a 16-bit word to two bytes (hi, lo)
func unword(w uint16) (uint8, uint8) {
	return uint8(w >> 8), uint8(w)
}

func hi(v uint16) uint8 {
	return uint8(v >> 8)
}

func lo(v uint16) uint8 {
	return uint8(v)
}
