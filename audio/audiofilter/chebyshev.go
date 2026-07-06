package audiofilter

// A Chebyshev holds the state for a 3rd order chebyshev type 1 filter.
type Chebyshev struct {
	xv [4]float64
	yv [4]float64
}

// Apply implements the Filter interface for a low-pass filter with
// cut off at 8700Hz for 32kHz sample rate.
func (f *Chebyshev) Apply(v float64) float64 {
	// magic values derived using http://jaggedplanet.com/iir/iir-explorer.asp
	const (
		gain = 6.9688733859366865 * 4
		a0   = -0.18936162546121998
		a1   = 0.57858619908831900
		a2   = -0.24126283620421146
		b0   = 1.0
		b1   = 3.0
		b2   = 3.0
		b3   = 1.0
	)

	f.xv[0] = f.xv[1]
	f.xv[1] = f.xv[2]
	f.xv[2] = f.xv[3]
	f.yv[0] = f.yv[1]
	f.yv[1] = f.yv[2]
	f.yv[2] = f.yv[3]

	f.xv[3] = v / gain
	f.yv[3] = 0 +
		b0*f.xv[0] - a0*f.yv[0] +
		b1*f.xv[1] - a1*f.yv[1] +
		b2*f.xv[2] - a2*f.yv[2] +
		b3*f.xv[3]

	return f.yv[3]
}
