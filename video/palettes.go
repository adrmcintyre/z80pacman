package video

// A Palette identifies a colour palette (a 3 entry lookup table)
type Palette byte

// A paletteEntry is a sequence of 4 bytes representing a 4-colour lookup table.
//
// Sprites and tiles are encoded as 2-bits-per-pixel bitmaps, with each possible
// value of 00, 01, 10, 11 being rendered as a colour drawn from a specified
// palette. Each location in a paletteEntry is actually a 4-bit index into the
// separate colorData table. Color 00 in the bitmaps is reserved as "transparent",
// and never interpreted in terms of colorData.
type paletteEntry [4]byte

// Format: color_index0, color_index1, color_index2, color_index3
var paletteData = [64]paletteEntry{
}
