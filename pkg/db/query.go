package db

import (
	"context"
	"fmt"

	"github.com/scalarorg/scalar-service/config"
	"gorm.io/gorm"
)

func BuildTokenSentsBaseQuery(db *gorm.DB, extendWhereClause func(db *gorm.DB)) *gorm.DB {
	query := db.Table("token_sents ts").
		Select(`
            ts.*,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
			to_timestamp(sbh.block_time) as source_created_at,
            to_timestamp(dbh.block_time) as executed_created_at
        `)
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON ts.tx_hash = ce.command_id").
		Joins("LEFT JOIN block_headers sbh ON ts.source_chain = sbh.chain AND ts.block_number = sbh.block_number").
		Joins("LEFT JOIN block_headers dbh ON ce.source_chain = dbh.chain AND ce.block_number = dbh.block_number")
}

func BuildContractCallWithTokenBaseQuery(db *gorm.DB, extendWhereClause func(db *gorm.DB)) *gorm.DB {
	query := db.Table("contract_call_with_tokens ccwtk").
		Select(`
            ccwtk.*,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
            ce.created_at as executed_created_at
        `)
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN contract_call_approved_with_mints ccawm ON ccwtk.event_id = ccawm.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON ccawm.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON ccwtk.tx_hash = ce.command_id")
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

func GetTransferTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	var result BaseCrossChainTxResult

	query := BuildTokenSentsBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.Where("ts.tx_hash = ?", txHash).First(&result).Error

	return &result, err
}

func ListBridgeTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	query := BuildTokenSentsBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ts.source_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateCrossChainTxs(ctx, query, size, offset)
}

func GetBridgeTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	var result BaseCrossChainTxResult

	query := BuildTokenSentsBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ts.source_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	fmt.Println("query", query)

	err := query.Where("ts.tx_hash = ?", txHash).First(&result).Error

	return &result, err
}

func ListRedeemTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	query := DB.Relayer.Table("contract_call_with_tokens ccwtk").
		Select(`
            ccwtk.*,
            ccwtk.tx_hash as command_id,
            rdt.tx_hash as executed_tx_hash,
            rdt.block_number as executed_block_number,
            rdt.custodian_group_uid as executed_address,
			to_timestamp(sbh.block_time) as source_created_at,
            to_timestamp(rdt.block_time) as executed_created_at
        `).Where("ccwtk.destination_chain = ?", config.Env.BITCOIN_CHAIN_ID).
		Joins("LEFT JOIN redeem_txes rdt ON ccwtk.custodian_group_uid = rdt.custodian_group_uid AND rdt.session_sequence = rdt.session_sequence").
		Joins("LEFT JOIN block_headers sbh ON ccwtk.source_chain = sbh.chain AND ccwtk.block_number = sbh.block_number")

	return AggregateCrossChainTxs(ctx, query, size, offset)
}

func GetRedeemTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	var result BaseCrossChainTxResult

	query := BuildContractCallWithTokenBaseQuery(DB.Relayer, func(db *gorm.DB) {
		db.Where("ccwtk.destination_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.Where("ccwtk.tx_hash = ?", txHash).First(&result).Error

	return &result, err
}
