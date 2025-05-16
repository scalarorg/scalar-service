package db

import (
	"fmt"

	"github.com/scalarorg/scalar-service/pkg/types"
)

func StatTransactionBySourceChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			source_chain as chain,
			COUNT(*) as amount
		FROM token_sents
		GROUP BY chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Relayer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}

func StatTransactionByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			destination_chain as chain,
			COUNT(*) as amount
		FROM token_sents
		GROUP BY destination_chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Relayer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}

func StatTransactionByPath(limit int) ([]*types.PathAmount, error) {
	var stats []*types.PathAmount
	query := `
		SELECT 
			source_chain,
			destination_chain,
			COUNT(*) as amount
		FROM token_sents
		GROUP BY source_chain, destination_chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Relayer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}
