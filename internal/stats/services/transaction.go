package services

import (
	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/types"
)

func StatTransactionBySourceChain(limit int) ([]*types.ChainAmount, error) {
	return db.StatTransactionBySourceChain(limit)
}

func StatTransactionByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	stats, err := db.StatTransactionByDestinationChain(limit)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func StatTransactionByPath(limit int) ([]*types.PathAmount, error) {
	stats, err := db.StatTransactionByPath(limit)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
