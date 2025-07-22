package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/scalarorg/scalar-service/pkg/types"
)

func StatTransactionBySourceChain(limit int) ([]*types.ChainAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stats []*types.ChainAmount

	// Optimized query with proper indexing
	query := `
		SELECT 
			chain,
			COUNT(*) as amount
		FROM vault_transactions
		WHERE chain IS NOT NULL
			AND chain != ''
			AND TRIM(chain) != ''
		GROUP BY chain
		ORDER BY amount DESC
		LIMIT $1
	`

	err := DB.Indexer.WithContext(ctx).Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction stats by source chain: %w", err)
	}

	// Optimize chain name formatting for consistency
	for i := range stats {
		// Handle different chain formats (bitcoin|4, evm|ethereum, etc.)
		if !strings.Contains(stats[i].Chain, "|") {
			stats[i].Chain = "evm|" + stats[i].Chain
		}
	}
	return stats, nil
}

func StatTransactionByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stats []*types.ChainAmount

	// Optimized query with proper filtering
	query := `
		SELECT 
			destination_chain as chain,
			COUNT(*) as amount
		FROM vault_transactions
		GROUP BY destination_chain
		ORDER BY amount DESC
		LIMIT $1
	`

	err := DB.Indexer.WithContext(ctx).Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction stats by destination chain: %w", err)
	}

	// Optimize chain name formatting for consistency
	for i := range stats {
		// Handle different chain formats (bitcoin|4, evm|ethereum, etc.)
		if !strings.Contains(stats[i].Chain, "|") {
			stats[i].Chain = "evm|" + stats[i].Chain
		}
	}
	return stats, nil
}

func StatTransactionByPath(limit int) ([]*types.PathAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var stats []*types.PathAmount

	// Optimized query with proper filtering and composite index usage
	// Use direct column names without COALESCE to avoid type conversion issues
	query := `
		SELECT 
		chain as source_chain,
		destination_chain,
		COUNT(*) as amount
		FROM vault_transactions
		GROUP BY source_chain, destination_chain
		ORDER BY amount DESC
		LIMIT $1
	`

	err := DB.Indexer.WithContext(ctx).Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction stats by path: %w", err)
	}

	// Optimize chain name formatting for consistency
	for i := range stats {
		// Handle different chain formats for source chain
		if !strings.Contains(stats[i].SourceChain, "|") {
			stats[i].SourceChain = "evm|" + stats[i].SourceChain
		}
		// Handle different chain formats for destination chain
		if !strings.Contains(stats[i].DestinationChain, "|") {
			stats[i].DestinationChain = "evm|" + stats[i].DestinationChain
		}
	}
	return stats, nil
}
