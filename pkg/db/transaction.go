package db

import (
	"fmt"
	"strings"

	"github.com/scalarorg/scalar-service/pkg/types"
)

func StatTransactionBySourceChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			chain,
			COUNT(*) as amount
		FROM vault_transactions
		GROUP BY chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, limit).Scan(&stats).Error
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
		FROM vault_transactions
		GROUP BY destination_chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	for _, stat := range stats {
		if !strings.HasPrefix(stat.Chain, "evm|") {
			stat.Chain = "evm|" + stat.Chain
		}
	}
	return stats, nil
}

func StatTransactionByPath(limit int) ([]*types.PathAmount, error) {
	var stats []*types.PathAmount
	query := `
		SELECT 
			chain as source_chain,
			destination_chain,
			COUNT(*) as amount
		FROM vault_transactions
		GROUP BY source_chain, destination_chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	for _, stat := range stats {
		if !strings.HasPrefix(stat.DestinationChain, "evm|") {
			stat.DestinationChain = "evm|" + stat.DestinationChain
		}
	}
	return stats, nil
}
