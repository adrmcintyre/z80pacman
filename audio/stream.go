package audio

import (
	"io"
	"time"

	audiofilter "github.com/adrmcintyre/z80/audio/audiofilter"
	ebiten_audio "github.com/hajimehoshi/ebiten/v2/audio"
)

// Specifies the audio system's sample frequency and buffer size.
// Higher sample frequency requires a shorter buffer size
// If buffer is too long, audio will lag the action.
// If buffer is too short, the audio becomes choppy.
type Latency struct {
	sampleRate int64
	bufferSize time.Duration
}

var (
	// LatencyLow is suitable for desktop builds
	LatencyLow = Latency{96000, 16055123 * time.Nanosecond}
	// LatencyHigh is suitable for wasm builds
	LatencyHigh = Latency{18000, 120 * time.Millisecond}
)

// Configure some parameters of the simulated hardware
const (
	voiceCount = 3              // how many voices are supported
	zeroOutput = uint16(0x0000) // value of zero-output
)

// An hwVoice represents the current state of the registers for 1 voice.
type hwVoice struct {
	wave byte   // low 3 bits used – selects waveform 0-7 from ROM
	vol  byte   // low nibble – 0 off to 15 loudest
	freq uint32 // real hardware has 20 bits for voice 0; 16 bits voices 1, 2
}

// Connect connects an ebiten audio player to
// the output of the simulated hardware on the host.
func (au *Audio) Connect(latency Latency) error {
	audioContext := ebiten_audio.NewContext(int(latency.sampleRate))
	audioPlayer, err := audioContext.NewPlayer(au)
	if err != nil {
		return err
	}
	au.player = audioPlayer
	au.sampleRate = latency.sampleRate
	au.filter = audiofilter.Compose{
		&audiofilter.ExpMovingAvg{},
		&audiofilter.Chebyshev{},
	}
	au.player.SetBufferSize(latency.bufferSize)
	au.player.Play()
	return nil
}

// Read is io.Reader's Read.
//
// Read fills buf with sampled audio according to hwVoice settings.
func (au *Audio) Read(buf []byte) (int, error) {
	const (
		bytesPerValue  = 2
		bytesPerSample = bytesPerValue * 2 // 2 x 16-bit samples (for left and right)
	)

	sampleRate := float64(au.sampleRate)
	alignedLen := len(buf) / bytesPerSample * bytesPerSample

	numSamples := alignedLen / bytesPerSample
	numEmitted := au.pos / bytesPerSample

	t0 := time.Now().Add(queueDelay)
	if !enableQueue {
		loadRegisters()
	}

	enabled := soundEnabled.Load()

	// sample and mix channels
	for i := range numSamples {
		sampleIndex := numEmitted + int64(i)
		t := float64(sampleIndex) / sampleRate

		if enableQueue {
			processWriteQueue(t0.Add(time.Microsecond * time.Duration(i) / time.Duration(sampleRate) / 1e6))
		}

		var value uint16
		if enabled {
			var j int64
			for ch, channel := range voice {
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

		output := zeroOutput + uint16(au.filter.Apply(float64(value)))

		// encode left channel
		buf[4*i] = byte(output)
		buf[4*i+1] = byte(output >> 8)
		// same audio on the right channel
		buf[4*i+2] = byte(output)
		buf[4*i+3] = byte(output >> 8)
	}

	au.pos += int64(alignedLen)

	return alignedLen, nil
}

// Close is io.Closer's Close.
func (au *Audio) Close() error {
	_ = au.player.Close()
	return nil
}

// assert we implement the interface
var _ io.Closer = (*Audio)(nil)
