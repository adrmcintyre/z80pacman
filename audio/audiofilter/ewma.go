package audiofilter

// An ExpMovingAvg holds the state for an exponential moving average
// filter.
type ExpMovingAvg struct {
	xv float64
	yv float64
}

// Apply implements the Filter interface for a low-pass moving exponential
// average filter.
func (f *ExpMovingAvg) Apply(x float64) float64 {
	const alpha = 0.95
	y := (f.xv+x)/2*(1-alpha) + (f.yv)*alpha
	f.xv = x
	f.yv = y
	return y * 2
}
