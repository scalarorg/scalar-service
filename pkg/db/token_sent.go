package db

import (
	"context"
	"fmt"
	"time"

	"github.com/scalarorg/bitcoin-vault/go-utils/chain"
	"github.com/scalarorg/data-models/chains"
	"github.com/scalarorg/scalar-service/config"
	"gorm.io/gorm"
)

func BuildTokenSentBaseQuery(db *gorm.DB, where func(db *gorm.DB) *gorm.DB) *gorm.DB {
	query := db.Table("token_sents ts").
		Select(`
            ts.*,
            ce.command_id,
            ce.tx_hash as executed_tx_hash,
            ce.block_number as executed_block_number,
            ce.address as executed_address,
            ce.created_at as executed_created_at
        `)
	query = where(query)
	return query.Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
		Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")
}

type TokenSentsResult struct {
	// TokenSent fields
	EventID              string `gorm:"column:event_id"`
	TxHash               string `gorm:"column:tx_hash"`
	BlockNumber          uint64 `gorm:"column:block_number"`
	SourceChain          string `gorm:"column:source_chain"`
	SourceAddress        string `gorm:"column:source_address"`
	DestinationChain     string `gorm:"column:destination_chain"`
	DestinationAddress   string `gorm:"column:destination_address"`
	TokenContractAddress string `gorm:"column:token_contract_address"`
	Amount               uint64 `gorm:"column:amount"`
	Symbol               string `gorm:"column:symbol"`
	Status               string `gorm:"column:status"`

	// TokenSentApproved specific fields
	CommandID string `gorm:"column:command_id"`

	// CommandExecuted specific fields
	ExecutedTxHash      string `gorm:"column:executed_tx_hash"`
	ExecutedBlockNumber uint64 `gorm:"column:executed_block_number"`
	ExecutedAddress     string `gorm:"column:executed_address"`

	CreatedAt                time.Time `gorm:"column:created_at"`
	ExecutedCommandCreatedAt time.Time `gorm:"column:executed_created_at"`
}

var _ (ExpectedCrossChainDocument) = (*TokenSentsResult)(nil)

func (b *TokenSentsResult) GetID() string {
	return b.EventID
}

func (b *TokenSentsResult) GetType() CrossChainTx {
	typ := CrossChainTxBridge
	if b.SourceChain != config.Env.BITCOIN_CHAIN_ID {
		typ = CrossChainTxTransfer
	}
	return typ
}

func (b *TokenSentsResult) GetStatus() string {
	return b.Status
}

func (b *TokenSentsResult) GetSource() *SourceDocument {
	var c = &chain.ChainInfo{}
	var name = ""
	err := c.FromString(b.DestinationChain)
	if err != nil {
		name = b.DestinationChain
	} else {
		name = chain.GetDisplayedName(*c)
	}
	return &SourceDocument{
		BaseDocument: &BaseDocument{
			Chain:     b.SourceChain,
			ChainName: name,
			TxHash:    b.TxHash,
			Status:    b.Status,
			Value:     fmt.Sprintf("%d", b.Amount),
			Fee:       "0",
			// TODO: fix token
			CrossChainAsset: CrossChainAsset{
				// TODO: queyr in chains
				Name:     b.Symbol,
				Symbol:   b.Symbol,
				Address:  b.TokenContractAddress,
				Decimals: 8,
				IsNative: false,
			},
			BlockHeight: b.BlockNumber,
			// TODO: refactor time mechanism
			CreatedAt: uint64(b.CreatedAt.Unix()),
		},
		Sender: b.SourceAddress,
	}
}

func (b *TokenSentsResult) GetDestination() *DestinationDocument {
	var c = &chain.ChainInfo{}
	var name = ""
	err := c.FromString(b.DestinationChain)
	if err != nil {
		name = b.DestinationChain
	} else {
		name = chain.GetDisplayedName(*c)
	}

	var status = chains.TokenSentStatusPending
	if b.TxHash != "" && b.ExecutedTxHash != "" {
		status = chains.TokenSentStatusSuccess
	}

	return &DestinationDocument{
		BaseDocument: &BaseDocument{
			Chain:     b.DestinationChain,
			ChainName: name,
			TxHash:    b.ExecutedTxHash,

			// TODO: use token approved status or executed status
			Status: string(status),
			Value:  fmt.Sprintf("%d", b.Amount),
			Fee:    "0",
			CrossChainAsset: CrossChainAsset{
				// TODO: query in chains to get token info
				Name:     b.Symbol,
				Symbol:   b.Symbol,
				Address:  b.TokenContractAddress,
				Decimals: 8,
				IsNative: false,
			},
			BlockHeight: b.ExecutedBlockNumber,
			// TODO: refactor time mechanism
			CreatedAt: uint64(b.ExecutedCommandCreatedAt.Unix()),
		},
		Receiver: b.DestinationAddress,
	}
}

func (b *TokenSentsResult) GetCommandID() string {
	return b.CommandID
}

func AggregateTokenSents(ctx context.Context, query *gorm.DB, size, offset int) ([]TokenSentsResult, int, error) {
	var results []TokenSentsResult

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

func ListTransferTxs(ctx context.Context, size, offset int) ([]TokenSentsResult, int, error) {
	query := BuildTokenSentBaseQuery(DB.Relayer, func(db *gorm.DB) *gorm.DB {
		return db.Where("ts.source_chain <> ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateTokenSents(ctx, query, size, offset)
}

func ListBridgeTxs(ctx context.Context, size, offset int) ([]TokenSentsResult, int, error) {
	query := BuildTokenSentBaseQuery(DB.Relayer, func(db *gorm.DB) *gorm.DB {
		return db.Where("ts.source_chain = ?", config.Env.BITCOIN_CHAIN_ID)
	})

	return AggregateTokenSents(ctx, query, size, offset)
}
