package main

import (
	"github.com/adrmcintyre/z80pacman/audio"
)

var (
	org uint16 // address of next byte to assemble
)

// asm inserts 1 or more bytes into the program.
func asm(op uint8, args ...uint8) {
	programROM[org] = op
	org++
	for _, arg := range args {
		programROM[org] = arg
		org++
	}
}

// rel returns a 1-byte relative reference to an address,
// suitable for assembling relative jumps.
func rel(dst uint16) uint8 {
	return uint8(dst - (org + 2))
}

func lo(v uint16) uint8 {
	return uint8(v)
}

func hi(v uint16) uint8 {
	return uint8(v >> 8)
}

// loadTestProgram loads the specified test program.
func loadTestProgram(prog string) {
	// preamble to disable interrupts and then jump to
	// selected program at 0x0100
	org = 0x0000
	asm(0xf3)             // di
	asm(0x3e, 0x00)       // ld a, #00
	asm(0xed, 0x47)       // ld i, a
	asm(0xc3, 0x00, 0x01) // jp #0100

	org = 0x100

	switch prog {
	case "1":
		test1()
	case "2":
		test2()
	default:
		die("unknown program: %s", prog)
	}

	// if we reach here, halt and dump
	setBreakpoint(org, Breakpoint{
		Resume:      false,
		DumpProgram: true,
	})
	asm(0x76) // hlt
}

// test1 plays a chord (middle C, E, G) on the three audio channels.
func test1() {
	audio.SetSoundEnable(true)

	freq0 := (440 * 4096 / 375)
	freq1 := (523 * 4096 / 375)
	freq2 := (660 * 4096 / 375)

	// channel 0
	asm(0xdd, 0x21, 0x40, 0x50) // ld ix, #5040 // acc/wave, followed by freq/vol

	asm(0xdd, 0x36, 0x10, uint8(freq0>>0))  // ld (ix+off), #val
	asm(0xdd, 0x36, 0x11, uint8(freq0>>4))  // ld (ix+off), #val
	asm(0xdd, 0x36, 0x12, uint8(freq0>>8))  // ld (ix+off), #val
	asm(0xdd, 0x36, 0x13, uint8(freq0>>12)) // ld (ix+off), #val
	asm(0xdd, 0x36, 0x14, uint8(freq0>>16)) // ld (ix+off), #val
	asm(0xdd, 0x36, 0x15, 0x0f)             // max volume

	// channel 1
	asm(0xdd, 0x36, 0x16, uint8(freq1>>4))  // freq
	asm(0xdd, 0x36, 0x17, uint8(freq1>>8))  // freq
	asm(0xdd, 0x36, 0x18, uint8(freq1>>12)) // freq
	asm(0xdd, 0x36, 0x19, uint8(freq1>>16)) // freq
	asm(0xdd, 0x36, 0x1a, 0x0f)             // max volume

	// channel 2
	asm(0xdd, 0x36, 0x0f, 0x00)             // voice
	asm(0xdd, 0x36, 0x1b, uint8(freq2>>4))  // freq
	asm(0xdd, 0x36, 0x1c, uint8(freq2>>8))  // freq
	asm(0xdd, 0x36, 0x1d, uint8(freq2>>12)) // freq
	asm(0xdd, 0x36, 0x1e, uint8(freq2>>16)) // freq
	asm(0xdd, 0x36, 0x1f, 0x0f)             // max volume
}

// test2 plays a repeated ascending tone on the second audio channel.
func test2() {
	audio.SetSoundEnable(true)
	// labels
	var (
		next_freq uint16
		got_frame uint16
		halt_lp   uint16
	)

	const (
		irq_vec       uint16 = 0x00fe
		prog_start    uint16 = 0x1000
		stack_top     uint16 = 0x4fc0
		io_irq_en     uint16 = 0x5000
		io_audio_base uint16 = 0x5040
	)

	const (
		freq0  uint16 = 440 * 4096 / 375
		delta0 uint16 = 100

		nth_frame uint8 = 5
		cnt       uint8 = 40
	)

	asm(0x31, lo(stack_top), hi(stack_top)) // ld sp, #stack_top
	asm(0xed, 0x5e)                         // im 2 - interrupt mode 2
	asm(0x3e, hi(irq_vec))                  // ld a, #00
	asm(0xed, 0x47)                         // ld i, a // vector via 00..
	asm(0x3e, lo(irq_vec))                  // ld a, #fe
	asm(0xd3, 0x00)                         // out (0), a	// vector via 00fe
	asm(0x3e, 0x01)                         // ld a, #01
	asm(0x32, lo(io_irq_en), hi(io_irq_en)) // ld (#5000),a ; enable interrupt hardware

	asm(0x21, lo(freq0), hi(freq0))   // ld hl, #freq0
	asm(0x11, lo(delta0), hi(delta0)) // ld de, #delta0
	asm(0x0e, nth_frame)              // ld c, #nth_frame
	asm(0x06, cnt)                    // ld b, #cnt
	asm(0xfb)                         // ei ; enable interrupts

	halt_lp = org
	asm(0x76)               // hlt ; wait for interrupt
	asm(0x18, rel(halt_lp)) // jr halt_lp

	org = irq_vec
	asm(lo(prog_start), hi(prog_start)) // point interrupt vector at 0x1000

	for range 2 {
		org = prog_start

		asm(0xdd, 0x21, lo(io_audio_base), hi(io_audio_base)) // ld ix, #5040 // acc/wave, followed by freq/vol

		asm(0x0d)                 // dec c
		asm(0x28, rel(got_frame)) // jr z, ...
		asm(0xfb)                 // ei
		asm(0xc9)                 // ret

		got_frame = org
		asm(0x0e, nth_frame)            // ld c, #nth_frame
		asm(0x05)                       // dec b
		asm(0x20, rel(next_freq))       // jr nz, next_freq
		asm(0x21, lo(freq0), hi(freq0)) // ld hl, #freq0
		asm(0x06, cnt)                  // ld b, #cnt

		next_freq = org
		asm(0x19)                   // add hl, de
		asm(0x7d)                   // ld a,l
		asm(0xdd, 0x77, 0x10)       // ld (ix+#10), a
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0xdd, 0x77, 0x11)       // ld (ix+#11), a
		asm(0x7c)                   // ld a,h
		asm(0xdd, 0x77, 0x12)       // ld (ix+#12), a
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0x0f)                   // rrca
		asm(0xdd, 0x77, 0x13)       // ld (ix+#13), a
		asm(0xdd, 0x36, 0x15, 0x0f) // ld (ix+#15), #15 ; max volume

		setBreakpoint(org, Breakpoint{
			DumpHisto: true,
		})
		asm(0xfb) // ei
		asm(0xc9) // ret
	}
}
