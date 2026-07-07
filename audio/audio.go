package audio

import (
	"sync"
	"sync/atomic"
)

const (
	voiceCount = 3 // how many simulated voices are supported
)

var (
	regMutex sync.Mutex
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

	soundEnabled atomic.Bool // is audio enabled?
)

func SetSoundEnable(value bool) {
	soundEnabled.Store(value)
}

func AccWaveWrite(i uint16, value uint8) {
	regMutex.Lock()
	defer regMutex.Unlock()

	accWaveReg[i] = value
}

func FreqVolWrite(i uint16, value uint8) {
	regMutex.Lock()
	defer regMutex.Unlock()

	freqVolReg[i] = value
}
