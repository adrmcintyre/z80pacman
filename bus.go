package main

import (
	"github.com/adrmcintyre/z80pacman/audio"
	"github.com/adrmcintyre/z80pacman/video"
)

var (
	programROM [0x4000]uint8
	programRAM [0x400]uint8
)

// ROM appears at 0x0000..0x3fff
// RAM appears at 0x4000..0x4fff and again from 0x6000
// I/O appears at 0x5000..0x5fff and again from 0x7000, and it repeats every 0x0100 within this regions.
// Entire memory maps repeats again from 0x8000.

// Note program RAM at 0x4ff0..0x4fff is read directly by the video
// hardware to get the sprite number and x/y flip bits. This is
// controlled by the U6,U5,U4 muxes in the sync bus controller.

func busRead(addr uint16) uint8 {
	var v uint8

	// bit15 is ignored, so entire memory map repeats at 0x8000
	addr &^= 0x8000

	switch addr & 0x4000 {
	case 0x0000:
		// 0000..3fff - ROM
		// early exit to skip memory tracing for ROM
		return programROM[addr]

	case 0x4000:
		// 4000..5fff - RAM and I/O; notice bit 13 not used, hence the repeat at 0x6000
		switch addr & 0x1000 {
		case 0x0000:
			// 4000..4fff - RAM
			switch addr & 0x0c00 {
			case 0x0000:
				// 4000..43ff - Tile RAM
				v = uint8(video.TileRAM[addr&0x3ff])
			case 0x0400:
				// 4400..47ff - Palette RAM
				v = uint8(video.PalRAM[addr&0x3ff])
			case 0x0800:
				// 4800..4bff - RAM absent
				v = 0
			case 0x0c00:
				// 4c00..4fef - Program RAM
				v = programRAM[addr&0x3ff]
			}

		case 0x1000:
			// 5000..5fff - I/O; notice bits 11..8 not used, hence repeat every 0x0100
			switch addr & 0x00c0 {
			case 0x0000:
				// 5000..503f - IN0
				v = uint8(in0_State.Load())

			case 0x0040:
				// 5040..507f - IN1
				v = uint8(in1_State.Load())

			case 0x0080:
				// 5080..50bf - DIPs
				v = uint8(dipSwitchState.Load())

			case 0x00c0:
				// 50c0..50ff - unused
				v = 0
			}
		}
	}

	debugTraceRead(addr, v)

	return v
}

func busWrite(addr uint16, data uint8) {
	addr &^= 0x8000

	debugTraceWrite(addr, data)

	switch addr & 0x4000 {
	case 0x0000:
		// 0000..3fff - ROM
	case 0x4000:
		// 4000..5fff - RAM and I/O
		switch addr & 0x1000 {
		case 0x0000:
			// 4000..4fff - RAM
			switch addr & 0x0c00 {
			case 0x0000:
				// 4000.43ff - Tile RAM
				video.Mutex.Lock()
				video.TileRAM[addr&0x3ff] = video.Tile(data)
				video.Mutex.Unlock()
			case 0x0400:
				// 4400..47ff - Palette RAM
				video.Mutex.Lock()
				video.PalRAM[addr&0x3ff] = video.Palette(data)
				video.Mutex.Unlock()
			case 0x0800:
				// 4800..4bff - no RAM
			case 0x0c00:
				// 4c00..4fef - Program RAM
				programRAM[addr&0x3ff] = data
				if addr&0x3f0 == 0x3f0 {
					// 4ff0..4fff - Sprite look RAM
					video.Mutex.Lock()
					video.SpriteLookRAM[addr&0xf] = data
					video.Mutex.Unlock()
				}
			}

		case 0x1000:
			// 5000..5fff - I/O
			switch addr & 0x00c0 {
			case 0x0000:
				// 5000..503f - latch
				latchWrite(int(addr&0x07), (data&1) == 1)

			case 0x0040:
				// 5040..507f - audio/video
				switch addr & 0x30 {
				case 0x00:
					//5040..504f - audio accumulators and waveforms
					audio.AccWaveWrite(addr&0x0f, data&0x0f)
				case 0x10:
					//5050..505f - audio frequencies and volumes
					audio.FreqVolWrite(addr&0x0f, data&0x0f)
				case 0x20:
					//5060..506f - sprite positions
					video.Mutex.Lock()
					video.SpritePosRegister[addr&0xf] = data
					video.Mutex.Unlock()
				case 0x30:
					//5070..507f - unused
				}

			case 0x0080:
				// 5080..50bf - unused

			case 0x00c0:
				// 50c0..50ff - watchdog; any write resets it
				watchdogReset()
			}
		}
	}
}

func ioWrite(addr uint16, data uint8) {
	irqLowRegister.Store(uint32(data))
}
