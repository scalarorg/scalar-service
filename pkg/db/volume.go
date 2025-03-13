package db

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/pkg/types"
)

func GetTopTransferUsers(limit int) ([]types.AddressAmount, error) {
	var stats []types.AddressAmount
	query := `
		SELECT 
			sender as address,
			SUM(asset_amount) as amount
		FROM event_token_sents
		GROUP BY address
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, limit).Scan(&stats).Error
	log.Info().Msgf("top transfer users: %v", stats)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}

func GetTopBridgeUsers(sourceChain string, limit int) ([]*types.AddressAmount, error) {
	var stats []*types.AddressAmount
	query := `
		SELECT 
			sender as address,
			SUM(asset_amount) as amount
		FROM event_token_sents
		WHERE chain = ?
		GROUP BY address
		ORDER BY amount DESC
		LIMIT ?
	`
	err := DB.Indexer.Raw(query, sourceChain, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch top transfer users: %w", err)
	}
	return stats, nil
}

func StatVolumeBySourceChain(limit int) ([]*types.ChainAmount, error) {
	var stats []*types.ChainAmount
	query := `
		SELECT 
			chain,
			SUM(asset_amount) as amount
		FROM event_token_sents
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
			SUM(asset_amount) as amount
		FROM event_token_sents
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
			SUM(asset_amount) as amount
		FROM event_token_sents
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
