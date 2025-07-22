package db

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/pkg/types"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

// Get top users by volume with optimized parallel queries and context timeout
func GetTopTransferUsers(limit int) ([]types.AddressAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	wg := sync.WaitGroup{}
	wg.Add(2)
	var transferStats []types.AddressAmount
	var transferErr error
	
	go func() {
		defer wg.Done()
		// Optimized query with better indexing strategy and LIMIT push-down
		query := `
		SELECT 
			source_address as address,
			SUM(amount) as amount
		FROM token_sents
		WHERE source_chain LIKE 'evm|%'
			AND amount > 0
		GROUP BY source_address
		ORDER BY amount DESC
		LIMIT $1
		`
		transferErr = DB.Indexer.WithContext(ctx).Raw(query, limit*2).Scan(&transferStats).Error
		if transferErr != nil {
			log.Error().Err(transferErr).Msg("failed to fetch top transfer users")
		}
	}()
	
	var bridgeStats []types.AddressAmount
	var bridgeErr error
	
	go func() {
		defer wg.Done()
		// Optimized query with better performance and proper filtering
		query := `
		SELECT 
			staker_script_pubkey as address,
			SUM(amount) as amount
		FROM vault_transactions
		WHERE staker_script_pubkey IS NOT NULL
			AND staker_script_pubkey != ''
			AND amount > 0
		GROUP BY staker_script_pubkey
		ORDER BY amount DESC
		LIMIT $1
		`
		bridgeErr = DB.Indexer.WithContext(ctx).Raw(query, limit*2).Scan(&bridgeStats).Error
		if bridgeErr != nil {
			log.Error().Err(bridgeErr).Msg("failed to fetch top bridge users")
		}
	}()

	wg.Wait()
	
	// Handle errors
	if transferErr != nil && bridgeErr != nil {
		return nil, fmt.Errorf("both queries failed: transfer=%v, bridge=%v", transferErr, bridgeErr)
	}
	
	// Merge and sort results efficiently
	allStats := mergeSortedStats(transferStats, bridgeStats, func(a, b types.AddressAmount) int {
		if a.Amount > b.Amount {
			return 1
		} else if a.Amount < b.Amount {
			return -1
		}
		return 0
	}, func(a, b types.AddressAmount) types.AddressAmount {
		return types.AddressAmount{
			Address: a.Address,
			Amount:  a.Amount + b.Amount,
		}
	})

	// Return top results
	if len(allStats) > limit {
		return allStats[len(allStats)-limit:], nil
	}
	return allStats, nil
}

func GetTopBridgeUsers(sourceChain string, limit int) ([]*types.AddressAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	var stats []*types.AddressAmount
	
	// Optimized query with proper indexing and filtering
	query := `
		SELECT 
			staker_script_pubkey as address,
			SUM(amount) as amount
		FROM vault_transactions
		WHERE chain = $1
			AND staker_script_pubkey IS NOT NULL
			AND staker_script_pubkey != ''
			AND amount > 0
		GROUP BY staker_script_pubkey
		ORDER BY amount DESC
		LIMIT $2
	`
	
	err := DB.Indexer.WithContext(ctx).Raw(query, sourceChain, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top bridge users: %w", err)
	}

	// Convert script pubkeys to addresses in parallel for better performance
	if len(stats) > 0 {
		var wg sync.WaitGroup
		var mu sync.Mutex
		errorChan := make(chan error, len(stats))
		
		for i := range stats {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				address, err := utils.ScriptPubKeyToAddress(stats[idx].Address, sourceChain)
				if err != nil {
					errorChan <- fmt.Errorf("failed to convert pubkey to address: %w", err)
					return
				}
				mu.Lock()
				stats[idx].Address = address.String()
				mu.Unlock()
			}(i)
		}
		
		wg.Wait()
		close(errorChan)
		
		// Check for conversion errors
		if err := <-errorChan; err != nil {
			return nil, err
		}
	}
	
	return stats, nil
}

func StatVolumeBySourceChain(limit int) ([]*types.ChainAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var stats []*types.ChainAmount
	
	// Optimized query with proper filtering and indexing
	query := `
		SELECT 
			chain,
			SUM(amount) as amount
		FROM vault_transactions
		WHERE amount > 0
		GROUP BY chain
		ORDER BY amount DESC
		LIMIT $1
	`
	
	err := DB.Indexer.WithContext(ctx).Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch volume by source chain: %w", err)
	}
	return stats, nil
}

func StatVolumeByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var stats []*types.ChainAmount
	
	// Optimized query with proper filtering
	query := `
		SELECT 
			destination_chain as chain,
			SUM(amount) as amount
		FROM vault_transactions
		WHERE amount > 0
		GROUP BY destination_chain
		ORDER BY amount DESC
		LIMIT $1
	`
	
	err := DB.Indexer.WithContext(ctx).Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch volume by destination chain: %w", err)
	}
	
	// Optimize chain name formatting
	for i := range stats {
		if !strings.HasPrefix(stats[i].Chain, "evm|") {
			stats[i].Chain = "evm|" + stats[i].Chain
		}
	}
	return stats, nil
}

func StatVolumeByPath(limit int) ([]*types.PathAmount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var stats []*types.PathAmount
	
	// Optimized query with proper filtering and composite index usage
	query := `
		SELECT 
			chain as source_chain,
			destination_chain,
			SUM(amount) as amount
		FROM vault_transactions
		WHERE amount > 0
		GROUP BY chain, destination_chain
		ORDER BY amount DESC
		LIMIT $1
	`
	
	err := DB.Indexer.WithContext(ctx).Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch volume by path: %w", err)
	}
	
	// Optimize chain name formatting
	for i := range stats {
		if !strings.HasPrefix(stats[i].DestinationChain, "evm|") {
			stats[i].DestinationChain = "evm|" + stats[i].DestinationChain
		}
	}
	return stats, nil
}
