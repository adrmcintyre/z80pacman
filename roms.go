package main

import (
	"embed"
)

//go:embed roms/*
var romFS embed.FS

var (
	programROM [0x4000]uint8
)

type romSet map[string][]uint8

func (roms romSet) Load() {
	for path, rom := range roms {
		data, err := romFS.ReadFile(path)
		if err != nil {
			die("%v", err)
		}
		if got, expected := len(data), len(rom); got != expected {
			die("%s: expected %d bytes, but got %d bytes", path, expected, got)
		}
		copy(rom, data)
	}
}
