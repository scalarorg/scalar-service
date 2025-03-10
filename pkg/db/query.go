package db

import (
	"context"

	"github.com/scalarorg/scalar-service/config"
	"gorm.io/gorm"
)

func BuildTokenSentsBaseQuery(db *gorm.DB, where func(db *gorm.DB)) *gorm.DB {
	query := db.Table("token_sents ts").
		Select(`
            ts.*,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
            ce.created_at as executed_created_at
        `)
	where(query)
	return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
		Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
}

func BuildContractCallWithTokenBaseQuery(db *gorm.DB, where func(db *gorm.DB)) *gorm.DB {
	query := db.Table("contract_call_with_tokens ccwtk").
		Select(`
            ccwtk.*,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
            ce.created_at as executed_created_at
        `)
	where(query)
	return query.Joins("LEFT JOIN contract_call_approved_with_mints ccawm ON ccwtk.event_id = ccawm.event_id").
		Joins("LEFT JOIN command_executeds ce ON ccawm.command_id = ce.command_id")
}

func AggregateCrossChainTxs(ctx context.Context, query *gorm.DB, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	var results []BaseCrossChainTxResult

	var totalCount int64

	if size <= 0 {
		size = 10
	}
	if offset < 0 {
		offset = 0
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&results).Error

	if err != nil {
		return nil, 0, err
	}

	return results, int(totalCount), nil
}

func ListTransferTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	query := BuildTokenSentsBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateCrossChainTxs(ctx, query, size, offset)
}

func ListBridgeTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	query := BuildTokenSentsBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ts.source_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateCrossChainTxs(ctx, query, size, offset)
}

func ListRedeemTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	query := BuildContractCallWithTokenBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ccwtk.destination_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateCrossChainTxs(ctx, query, size, offset)
}
