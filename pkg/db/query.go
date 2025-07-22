package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/scalarorg/data-models/chains"
	"github.com/scalarorg/scalar-service/config"
	"github.com/scalarorg/scalar-service/pkg/utils"
	"gorm.io/gorm"
)

func BuildVaultTxsBaseQuery(db *gorm.DB, extendWhereClause func(db *gorm.DB)) *gorm.DB {
	query := db.Table("vault_transactions vt").
		Select(`
            vt.tx_hash,
            vt.block_number,
            vt.timestamp as block_time,
			vt.chain as source_chain,
            vt.staker_script_pubkey as source_address,
            vt.destination_chain,
			vt.destination_recipient_address as destination_address,
			vt.destination_token_address as token_contract_address,
			vt.amount,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
            to_timestamp(dbh.block_time) as executed_block_time
        `).
		Where("vt.timestamp IS NOT NULL AND vt.amount > 0")
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON vt.tx_hash = ce.command_id").
		Joins("LEFT JOIN block_headers dbh ON ce.source_chain = dbh.chain AND ce.block_number = dbh.block_number").Order("vt.timestamp DESC")
}

func BuildTokenSentsBaseQuery(db *gorm.DB, extendWhereClause func(db *gorm.DB)) *gorm.DB {
	query := db.Table("token_sents ts").
		Select(`
            ts.*,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
            to_timestamp(dbh.block_time) as executed_block_time,
			dbh.block_time AS block_time
        `).
		Where("ts.block_time IS NOT NULL AND ts.amount > 0")
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON ts.tx_hash = ce.command_id").
		Joins("LEFT JOIN block_headers dbh ON ts.source_chain = dbh.chain AND ts.block_number = dbh.block_number").Order("ts.timestamp DESC")
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
        `).
		Where("ccwtk.block_number IS NOT NULL AND ccwtk.amount > 0")
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN contract_call_approved_with_mints ccawm ON ccwtk.event_id = ccawm.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON ccawm.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON ccwtk.tx_hash = ce.command_id").Order("ccwtk.block_number DESC")
}

func AggregateCrossChainTxs(ctx context.Context, query *gorm.DB, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	var results []BaseCrossChainTxResult

	var totalCount int64

	// Add timeout to context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if size <= 0 {
		size = 10
	}
	if offset < 0 {
		offset = 0
	}

	if err := query.WithContext(ctxWithTimeout).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	err := query.WithContext(ctxWithTimeout).
		Order("block_number DESC").
		Offset(offset).
		Limit(size).
		Find(&results).Error

	if err != nil {
		return nil, 0, err
	}
	for i, result := range results {
		if result.ExecutedTxHash != "" {
			results[i].Status = string(chains.TokenSentStatusSuccess)
		}
	}
	return results, int(totalCount), nil
}

func ListTransferTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	// Add timeout to context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	query := BuildTokenSentsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateCrossChainTxs(ctxWithTimeout, query, size, offset)
}

func GetTransferTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	// Add timeout to context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var result BaseCrossChainTxResult

	query := BuildTokenSentsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.WithContext(ctxWithTimeout).
		Where("ts.tx_hash = ? AND ts.tx_hash IS NOT NULL AND ts.tx_hash != ''", txHash).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get transfer transaction: %w", err)
	}

	return &result, nil
}

// Get all bridge txs from indexer's vault_transactions table
func ListBridgeTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	query := BuildVaultTxsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("vt.chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})
	results, count, err := AggregateCrossChainTxs(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}

	for i, result := range results {
		address, err := utils.ScriptPubKeyToAddress(result.SourceAddress, result.SourceChain)
		if err != nil {
			log.Printf("error converting script pubkey to address: %s, %s, %v\n", result.SourceAddress, result.SourceChain, err)
		} else {
			results[i].SourceAddress = address.String()
		}
	}

	return results, count, nil
}

func GetBridgeTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	// Add timeout to context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var result BaseCrossChainTxResult

	query := BuildVaultTxsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("vt.chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.WithContext(ctxWithTimeout).
		Where("vt.tx_hash = ? AND vt.tx_hash IS NOT NULL AND vt.tx_hash != ''", txHash).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get bridge transaction: %w", err)
	}

	return &result, nil
}

func ListRedeemTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
	// Add timeout to context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	query := DB.Indexer.Table("evm_redeem_txes ert").
		Select(`
            ert.*,
            brt.tx_hash as command_id,
            brt.tx_hash as executed_tx_hash,
            brt.block_number as executed_block_number,
            brt.custodian_group_uid as executed_address,
			to_timestamp(dbh.block_time) as source_created_at,
            to_timestamp(brt.block_time) as executed_created_at,
			dbh.block_time AS block_time
        `).Where("ert.destination_chain = ?", config.Env.BITCOIN_CHAIN_ID).
		Joins("LEFT JOIN btc_redeem_txes brt ON ert.custodian_group_uid = brt.custodian_group_uid AND ert.session_sequence = brt.session_sequence").
		Joins("LEFT JOIN block_headers dbh ON ert.source_chain = dbh.chain AND ert.block_number = dbh.block_number")

	return AggregateCrossChainTxs(ctxWithTimeout, query, size, offset)
}

func GetRedeemTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	// Add timeout to context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var result BaseCrossChainTxResult

	query := BuildContractCallWithTokenBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("ccwtk.destination_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.WithContext(ctxWithTimeout).
		Where("ccwtk.tx_hash = ? AND ccwtk.tx_hash IS NOT NULL AND ccwtk.tx_hash != ''", txHash).
		First(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get redeem transaction: %w", err)
	}

	return &result, nil
}
