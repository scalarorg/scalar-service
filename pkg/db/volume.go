package db

import (
	"fmt"

	// Public key
	// Network params
	// Address utilities
	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/pkg/types"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

func GetTopTransferUsers(limit int) ([]types.AddressAmount, error) {
	var stats []types.AddressAmount
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
	err := DB.Relayer.Raw(query, limit).Scan(&stats).Error
	log.Info().Msgf("top transfer users: %v", stats)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
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
			source_chain as chain,
			SUM(amount) as amount
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

func StatVolumeByDestinationChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			destination_chain as chain,
			SUM(amount) as amount
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

func StatVolumeByPath(limit int) ([]*types.PathAmount, error) {
	var stats []*types.PathAmount
	query := `
		SELECT 
			source_chain,
			destination_chain,
			SUM(amount) as amount
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
