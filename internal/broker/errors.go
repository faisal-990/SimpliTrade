package broker

import "errors"

// ErrInvalidSide is returned when an order's side is neither buy nor sell.
var ErrInvalidSide = errors.New("broker: invalid order side")
