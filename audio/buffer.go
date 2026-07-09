package audio

import (
	"sync"
)

// NOTE: in the real thing, ROM 82s126.3m is used for sequencing
// the audio hardware's state machine. We do not attempt to emulate
// at that level of detail.

const (
	// TODO - we're rederiving the VBLANK period here
	clockFreq        = 18_432_000 / 3
	videoFrameWidth  = 384
	videoFrameHeight = 264
	clocksPerTick    = videoFrameWidth * videoFrameHeight

	samplesPerTick = ebitenSampleRate * clocksPerTick / clockFreq
	maxTickBacklog = 12
	preBufSize     = samplesPerTick * maxTickBacklog
)

var (
	preMutex      sync.Mutex
	preBuf        [preBufSize]uint16
	preEmitted    = int64(0)
	preWriteIndex = 0
	preReadIndex  = 0
	preWaiting    = 0
)

// A voice represents the current state of the registers for 1 voice.
type voice struct {
	wave byte   // low 3 bits used – selects waveform 0-7 from ROM
	vol  byte   // low nibble – 0=off to 15=loudest
	freq uint32 // real hardware has 20 bits for voice 0; 16 bits voices 1, 2
}

// Tick pre-generates 1/60s of audio to be picked up by the audio stream.
// This ensures that changes to audio parameters by the software are picked
// up at exactly the right moment, preventing audio chopiness.
func Tick() {
	voices := getVoices()

	preMutex.Lock()
	defer preMutex.Unlock()

	enabled := soundEnabled.Load()

	// sample and mix channels
	for i := range samplesPerTick {
		sampleIndex := preEmitted + int64(i)
		t := float64(sampleIndex) / float64(ebitenSampleRate)

		var value uint16
		if enabled {
			var j int64
			for ch, channel := range voices {
				freq := channel.freq * 3
				// channel 0 has more freq bits allocated
				if ch == 0 {
					j = int64(waveLength*float64(freq)/32*t) % waveLength
				} else {
					j = int64(waveLength*float64(freq)/2*t) % waveLength
				}
				value += scaledWaveData[channel.vol][channel.wave][j]
			}
		}

		preBuf[preWriteIndex+i] = uint16(filter.Apply(float64(value)))
	}

	preWriteIndex += samplesPerTick
	if preWriteIndex >= len(preBuf) {
		preWriteIndex = 0
	}
	preEmitted += samplesPerTick
	preWaiting += samplesPerTick
}

func getVoices() [3]voice {
	regMutex.Lock()
	defer regMutex.Unlock()

	return [3]voice{
		{
			freq: getFreq(0, 5),
			wave: byte(accWaveReg[5]),
			vol:  byte(freqVolReg[5]),
		},
		{
			freq: getFreq(6, 4),
			wave: byte(accWaveReg[10]),
			vol:  byte(freqVolReg[10]),
		},
		{
			freq: getFreq(11, 4),
			wave: byte(accWaveReg[15]),
			vol:  byte(freqVolReg[15]),
		},
	}
}

func getFreq(i, n int) uint32 {
	r := uint32(0)
	for j := range n {
		r |= uint32(freqVolReg[i+j]) << (4 * j)
	}
	return r
}
