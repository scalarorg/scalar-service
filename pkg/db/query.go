package db

import (
	"context"
)

func ListCrossChainTxs[T any](ctx context.Context, size, offset int) ([]T, int, error) {
	var txs []T
	var totalCount int64

	if size <= 0 {
		size = 10
	}
	if offset < 0 {
		offset = 0
	}

	var p = new(T)

	query := DB.Model(p)

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&txs).Error

	if err != nil {
		return nil, 0, err
	}

	return txs, int(totalCount), nil
}
