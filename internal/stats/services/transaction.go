package services

import (
	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/types"
)

func StatTransactionBySourceChain(limit int) ([]*types.ChainAmount, error) {
	return db.StatTransactionBySourceChain(limit)
}

func StatTransactionByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	return db.StatTransactionByDestinationChain(limit)
}

func StatTransactionByPath(limit int) ([]*types.PathAmount, error) {
	return db.StatTransactionByPath(limit)
}
