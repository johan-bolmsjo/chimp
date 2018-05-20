package chimp

import "fmt"

// Coord2D is a two dimensional coordinate.
type Coord2D struct {
	X, Y float32
}

func (c *Coord2D) String() string {
	return fmt.Sprintf("(%f, %f)", c.X, c.Y)
}
