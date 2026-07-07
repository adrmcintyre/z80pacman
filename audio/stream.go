package audio

import (
	"io"
	"time"

	audiofilter "github.com/adrmcintyre/z80pacman/audio/audiofilter"
	ebiten_audio "github.com/hajimehoshi/ebiten/v2/audio"
)

// Specifies the audio system's sample frequency and buffer size.
// Higher sample frequency requires a shorter buffer size
// If buffer is too long, audio will lag the action.
// If buffer is too short, the audio becomes choppy.
const (
	ebitenSampleRate     = 96_000
	ebitenBufferDuration = 100 * time.Millisecond
)

// playback state
var (
	player *ebiten_audio.Player
	filter audiofilter.Filter // output filter
)

// Stream implements the io.Reader interface necessary for ebiten to be able to
// stream from it.
type Stream struct{}

var stream *Stream

// Init constructs an initialised Audio struct.
func Init() error {
	stream = &Stream{}

	audioContext := ebiten_audio.NewContext(ebitenSampleRate)
	audioPlayer, err := audioContext.NewPlayer(stream)
	if err != nil {
		return err
	}
	player = audioPlayer
	filter = audiofilter.Compose{
		&audiofilter.ExpMovingAvg{},
		&audiofilter.Chebyshev{},
	}
	player.SetBufferSize(ebitenBufferDuration)
	player.Play()
	return nil
}

func Shutdown() {
	stream.Close()
}

// Read is io.Reader's Read.
//
// Read fills buf with pre-buffered audio generated during each call to Tick().
func (*Stream) Read(buf []byte) (int, error) {
	const (
		bytesPerValue  = 2
		bytesPerSample = bytesPerValue * 2 // 2 x 16-bit samples (for left and right)

	)
	preMutex.Lock()
	defer preMutex.Unlock()

	numSamples := len(buf) / bytesPerSample

	if preWaiting < numSamples {
		return 0, nil
	}

	index := preReadIndex

	// sample and mix channels
	for i := range numSamples {
		output := preBuf[index]
		index++
		if index >= preBufSize {
			index = 0
		}

		// encode left channel
		buf[4*i] = byte(output)
		buf[4*i+1] = byte(output >> 8)
		// same audio on the right channel
		buf[4*i+2] = byte(output)
		buf[4*i+3] = byte(output >> 8)
	}

	preWaiting -= numSamples
	preReadIndex = index

	return numSamples * bytesPerSample, nil
}

// Close is io.Closer's Close.
func (*Stream) Close() error {
	_ = player.Close()
	return nil
}

// assert we implement the interface
var _ io.Closer = (*Stream)(nil)
