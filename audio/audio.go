package audio

import (
	"sync"
	"sync/atomic"
	"time"
)

type writeRequest struct {
	ts  time.Time
	reg uint8
	val uint8
}

var (
	mutex sync.Mutex
	// 5040..504f
	// 0-4 voice1_acc
	// 5 voice1_wave
	// 6-9 voice2_acc
	// 10 voice2_wave
	// 11-14 voice3_acc
	// 15 voice3_wave
	accWaveReg [16]uint8

	// 5050..505f
	// 0-4 voice1_freq
	// 5 voice1_vol
	// 6-9 voice2_freq
	// 10 voice2_vol
	// 11-14 voice3_freq
	// 15 voice3_vol
	freqVolReg [16]uint8

	enableQueue = false
	queueDelay  = -500 * time.Millisecond
	nextWrite   uint64
	nextRead    uint64
	queueSize   uint64
	writeQueue  [4096]writeRequest
	indexMask   = uint64(4095)

	soundEnabled atomic.Bool // is audio enabled?

	// simulated hardware state
	voice [voiceCount]hwVoice // current state of the simulated hardware
)

// An hwVoice represents the current state of the registers for 1 voice.
type hwVoice struct {
	wave byte   // low 3 bits used – selects waveform 0-7 from ROM
	vol  byte   // low nibble – 0 off to 15 loudest
	freq uint32 // real hardware has 20 bits for voice 0; 16 bits voices 1, 2
}

func SetSoundEnable(value bool) {
	soundEnabled.Store(value)
}

func AccWaveWrite(i uint16, value uint8) {
	mutex.Lock()
	defer mutex.Unlock()

	if enableQueue {
		writeQueue[nextWrite&indexMask] = writeRequest{
			ts:  time.Now(),
			reg: uint8(i),
			val: value,
		}
		nextWrite++
		queueSize++
	} else {
		accWaveReg[i] = value
	}
}

func FreqVolWrite(i uint16, value uint8) {
	mutex.Lock()
	defer mutex.Unlock()

	if enableQueue {
		writeQueue[nextWrite&indexMask] = writeRequest{
			ts:  time.Now(),
			reg: uint8(i + 16),
			val: value,
		}
		nextWrite++
		queueSize++
	} else {
		freqVolReg[i] = value
	}
}

func processWriteQueue(deadline time.Time) {
	mutex.Lock()
	defer mutex.Unlock()

	dirty := false
	for queueSize > 0 {
		w := writeQueue[nextRead&indexMask]
		if w.ts.After(deadline) {
			break
		}
		if w.reg < 16 {
			accWaveReg[w.reg] = w.val
		} else {
			freqVolReg[w.reg-16] = w.val
		}
		dirty = true
		nextRead++
		queueSize--
	}
	if dirty {
		loadRegisters()
	}
}

func loadRegisters() {
	freq := func(i, n int) uint32 {
		r := uint32(0)
		for j := range n {
			r |= uint32(freqVolReg[i+j]) << (4 * j)
		}
		return r
	}

	//Ignore accumulators

	voice[0] = hwVoice{
		freq: freq(0, 5),
		wave: byte(accWaveReg[5]),
		vol:  byte(freqVolReg[5]),
	}
	voice[1] = hwVoice{
		freq: freq(6, 4),
		wave: byte(accWaveReg[10]),
		vol:  byte(freqVolReg[10]),
	}
	voice[2] = hwVoice{
		freq: freq(11, 4),
		wave: byte(accWaveReg[15]),
		vol:  byte(freqVolReg[15]),
	}
}
