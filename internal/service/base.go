package service

import (
	"context"
	"math"
)

type BaseCalculator struct{}

func NewBaseCalculator() *BaseCalculator { return &BaseCalculator{} }

func (c *BaseCalculator) Add(_ context.Context, a, b float64) (float64, error) {
	return round(a+b, 5), nil
}
func (c *BaseCalculator) Sub(_ context.Context, a, b float64) (float64, error) {
	return round(a-b, 5), nil
}
func (c *BaseCalculator) Mul(_ context.Context, a, b float64) (float64, error) {
	return round(a*b, 5), nil
}

func (c *BaseCalculator) Div(_ context.Context, a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return round(a/b, 5), nil
}

func round(v float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(v*pow) / pow
}
