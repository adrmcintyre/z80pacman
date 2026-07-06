package audiofilter

// Compose is a collection of filters composed in sequence.
type Compose []Filter

// Apply implements the Filter interface for a sequence of filters.
func (fs Compose) Apply(x float64) float64 {
	for _, f := range fs {
		x = f.Apply(x)
	}
	return x
}
