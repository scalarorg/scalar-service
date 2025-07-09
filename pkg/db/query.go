package db

import (
	"context"
	"fmt"
	"log"

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
        `)
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON vt.tx_hash = ce.command_id").
		Joins("LEFT JOIN block_headers dbh ON ce.source_chain = dbh.chain AND ce.block_number = dbh.block_number")
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
        `)
	extendWhereClause(query)
	// return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
	// 	Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
	return query.Joins("LEFT JOIN command_executeds ce ON ts.tx_hash = ce.command_id").
		Joins("LEFT JOIN block_headers dbh ON ts.source_chain = dbh.chain AND ts.block_number = dbh.block_number")
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
	query := BuildTokenSentsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateCrossChainTxs(ctx, query, size, offset)
}

func GetTransferTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	var result BaseCrossChainTxResult

	query := BuildTokenSentsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.Where("ts.tx_hash = ?", txHash).First(&result).Error

	return &result, err
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
	var result BaseCrossChainTxResult

	query := BuildVaultTxsBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("vt.chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	fmt.Println("query", query)

	err := query.Where("vt.tx_hash = ?", txHash).First(&result).Error

	return &result, err
}

func ListRedeemTxs(ctx context.Context, size, offset int) ([]BaseCrossChainTxResult, int, error) {
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

	return AggregateCrossChainTxs(ctx, query, size, offset)
}

func GetRedeemTx(ctx context.Context, txHash string) (*BaseCrossChainTxResult, error) {
	var result BaseCrossChainTxResult

	query := BuildContractCallWithTokenBaseQuery(DB.Indexer, func(db *gorm.DB) {
		db.Where("ccwtk.destination_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	err := query.Where("ccwtk.tx_hash = ?", txHash).First(&result).Error

	return &result, err
}
