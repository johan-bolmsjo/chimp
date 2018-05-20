package chimp

import "fmt"

// f32cival is a closed interval [a, b].
type f32cival struct {
	a, b float32
}

// clamp v to interval
func (r *f32cival) clamp(v float32) float32 {
	switch {
	case v < r.a:
		v = r.a
	case v > r.b:
		v = r.b
	}
	return v
}

// normalize v to the interval [0, 1] based on the allowed input value interval. I.e.
// clamping is applied before normalization.
func (r *f32cival) normalize(v float32) float32 {
	return (r.clamp(v) - r.a) / (r.b - r.a)
}

func (r *f32cival) String() string {
	return fmt.Sprintf("[%f, %f]", r.a, r.b)
}
