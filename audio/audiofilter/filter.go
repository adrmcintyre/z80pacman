package audiofilter

// A Filter produces filtered output from an input.
type Filter interface {
	Apply(v float64) float64
}
