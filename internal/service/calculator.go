package service

import (
	"context"
	"errors"
)

var ErrDivisionByZero = errors.New("division by zero")

type Calculator interface {
	Add(ctx context.Context, a, b float64) (float64, error)
	Sub(ctx context.Context, a, b float64) (float64, error)
	Mul(ctx context.Context, a, b float64) (float64, error)
	Div(ctx context.Context, a, b float64) (float64, error)
}
