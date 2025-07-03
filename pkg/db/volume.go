package db

import (
	"fmt"
	"sort"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/pkg/types"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

// Get top users by volume,
// volume is defined as sum of bridge amount and transfer amount
func GetTopTransferUsers(limit int) ([]types.AddressAmount, error) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	var transferStats []types.AddressAmount
	go func() {
		defer wg.Done()
		query := `
		SELECT 
			source_address as address,
			SUM(amount) as amount
		FROM token_sents
		WHERE source_chain LIKE 'evm|%'
		GROUP BY address
		ORDER BY amount DESC		
		LIMIT ?
	`
		err := DB.Indexer.Raw(query, limit).Scan(&transferStats).Error
		if err != nil {
			log.Error().Err(err).Msg("failed to fetch top transfer users")
		}
	}()
	var bridgeStats []types.AddressAmount
	go func() {
		defer wg.Done()
		query := `
		SELECT 
			destination_recipient_address as address,
			SUM(amount) as amount
		FROM vault_transactions
		GROUP BY address
		ORDER BY amount DESC
		LIMIT ?
	`
		err := DB.Indexer.Raw(query, limit).Scan(&bridgeStats).Error
		if err != nil {
			log.Error().Err(err).Msg("failed to fetch top bridge users")
		}
	}()
	wg.Wait()
	sort.Slice(transferStats, func(i, j int) bool {
		return transferStats[i].Amount > transferStats[j].Amount
	})
	sort.Slice(bridgeStats, func(i, j int) bool {
		return bridgeStats[i].Amount > bridgeStats[j].Amount
	})
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

	return allStats[len(allStats)-limit:], nil
}

func GetTopBridgeUsers(sourceChain string, limit int) ([]*types.AddressAmount, error) {
	var stats []*types.AddressAmount
	// query := `
	// 	SELECT
	// 		source_address as address,
	// 		SUM(amount) as amount
	// 	FROM token_sents
	// 	WHERE source_chain = ?
	// 	GROUP BY address
	// 	ORDER BY amount DESC
	// 	LIMIT ?
	// `
	// err := DB.Relayer.Raw(query, sourceChain, limit).Scan(&stats).Error
	query := `
		SELECT 
			staker_script_pubkey as address,
			SUM(amount) as amount
		FROM vault_transactions
		WHERE chain = ?
		GROUP BY address	
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, sourceChain, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}

	for _, stat := range stats {
		address, err := utils.ScriptPubKeyToAddress(stat.Address, sourceChain)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey to address: %w", err)
		}
		stat.Address = address.String()
	}
	return stats, nil
}

func StatVolumeBySourceChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			chain as chain,
			SUM(amount) as amount
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

func StatVolumeByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			destination_chain as chain,
			SUM(amount) as amount
		FROM vault_transactions
		GROUP BY destination_chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}

func StatVolumeByPath(limit int) ([]*types.PathAmount, error) {
	var stats []*types.PathAmount
	query := `
		SELECT 
			chain as source_chain,
			destination_chain,
			SUM(amount) as amount
		FROM vault_transactions
		GROUP BY source_chain, destination_chain
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}
