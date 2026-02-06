package store

import (
	"context"

	"calculator-service/internal/model"
)

type ResultStore interface {
	Add(ctx context.Context, e model.ResultEntry) error
	ListRecent(ctx context.Context, n int) ([]model.ResultEntry, error)
}
