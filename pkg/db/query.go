package db

import (
	"context"

	"github.com/scalarorg/data-models/chains"
)

func ListTokenSents(ctx context.Context, options *Options) ([]*chains.TokenSent, int, error) {
	var tokenSents []*chains.TokenSent
	var totalCount int64

	if options.Size <= 0 {
		options.Size = 10
	}
	if options.Offset < 0 {
		options.Offset = 0
	}

	query := DB.Model(&chains.TokenSent{})

	if options.TxHash != "" {
		query = query.Where("tx_hash = ?", options.TxHash)
	} else if options.EventId != "" {
		query = query.Where("event_id = ?", options.EventId)
	}

	if options.EventId == "" && options.TxHash == "" {
		if err := query.Count(&totalCount).Error; err != nil {
			return nil, 0, err
		}
	}

	err := query.
		Order("created_at DESC").
		Offset(options.Offset).
		Limit(options.Size).
		Find(&tokenSents).Error

	if err != nil {
		return nil, 0, err
	}

	// For filtered searches, use the length of results as the count
	if options.EventId != "" || options.TxHash != "" {
		totalCount = int64(len(tokenSents))
	}

	return tokenSents, int(totalCount), nil
}

func ListEvmToBTCTransfers(ctx context.Context, options *Options) ([]*chains.ContractCallWithToken, int, error) {
	var contractCallWithTokens []*chains.ContractCallWithToken
	var totalCount int64

	if options.Size <= 0 {
		options.Size = 10
	}
	if options.Offset < 0 {
		options.Offset = 0
	}

	query := DB.Model(&chains.ContractCallWithToken{})

	if options.TxHash != "" {
		query = query.Where("tx_hash = ?", options.TxHash)
	} else if options.EventId != "" {
		query = query.Where("event_id = ?", options.EventId)
	}

	// use LIKE to match evm and bitcoin
	query = query.Where("source_chain LIKE ?", "%evm%")
	query = query.Where("destination_chain LIKE ?", "%bitcoin%")

	if options.EventId == "" && options.TxHash == "" {
		if err := query.Count(&totalCount).Error; err != nil {
			return nil, 0, err
		}
	}

	err := query.
		Order("created_at DESC").
		Offset(options.Offset).
		Limit(options.Size).
		Find(&contractCallWithTokens).Error

	if err != nil {
		return nil, 0, err
	}

	// For filtered searches, use the length of results as the count
	if options.EventId != "" || options.TxHash != "" {
		totalCount = int64(len(contractCallWithTokens))
	}

	return contractCallWithTokens, int(totalCount), nil
}
