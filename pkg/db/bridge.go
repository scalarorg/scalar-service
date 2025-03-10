package db

import (
	"context"
	"fmt"
	"time"
)

type BridgeTxsResult struct {
	// TokenSent fields
	EventID              string `gorm:"column:event_id"`
	TxHash               string `gorm:"column:tx_hash"`
	BlockNumber          uint64 `gorm:"column:block_number"`
	LogIndex             uint   `gorm:"column:log_index"`
	SourceChain          string `gorm:"column:source_chain"`
	SourceAddress        string `gorm:"column:source_address"`
	DestinationChain     string `gorm:"column:destination_chain"`
	DestinationAddress   string `gorm:"column:destination_address"`
	TokenContractAddress string `gorm:"column:token_contract_address"`
	Amount               uint64 `gorm:"column:amount"`
	Symbol               string `gorm:"column:symbol"`
	Status               string `gorm:"column:status"`

	// TokenSentApproved specific fields
	CommandId  string `gorm:"column:command_id"`
	TransferID uint64 `gorm:"column:transfer_id"`

	// CommandExecuted specific fields
	ExecutedTxHash      string `gorm:"column:executed_tx_hash"`
	ExecutedBlockNumber uint64 `gorm:"column:executed_block_number"`
	ExecutedLogIndex    uint   `gorm:"column:executed_log_index"`
	ExecutedAddress     string `gorm:"column:executed_address"`

	CreatedAt                time.Time `gorm:"column:created_at"`
	ExecutedCommandCreatedAt time.Time `gorm:"column:executed_created_at"`
}

func ListBridgeTxs(ctx context.Context, size, offset int) ([]BridgeTxsResult, int, error) {
	var results []BridgeTxsResult

	var totalCount int64

	if size <= 0 {
		size = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := DB.Table("token_sents ts").
		Select(`
        ts.*,
        tsa.tx_hash as approved_tx_hash,
        tsa.log_index as approved_log_index,
        tsa.command_id,
        tsa.transfer_id,
        ce.tx_hash as executed_tx_hash,
        ce.block_number as executed_block_number,
        ce.log_index as executed_log_index,
        ce.address as executed_address,
		ce.created_at as executed_created_at
    `).
		Joins("LEFT JOIN token_sent_approveds tsa ON ts.event_id = tsa.event_id").
		Joins("LEFT JOIN command_executeds ce ON tsa.command_id = ce.command_id")

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

var _ (ExpectedCrossChainDocument) = (*BridgeTxsResult)(nil)

func (b *BridgeTxsResult) GetID() string {
	return b.EventID
}

func (b *BridgeTxsResult) GetType() CrossChainTx {
	return CrossChainTxBridge
}

func (b *BridgeTxsResult) GetStatus() string {
	return b.Status
}

func (b *BridgeTxsResult) GetSource() *SourceDocument {
	return &SourceDocument{
		BaseDocument: &BaseDocument{
			Chain: b.SourceChain,
			// TODO:
			ChainName: "b.ChainName",
			TxHash:    b.TxHash,

			Status: b.Status,
			Value:  fmt.Sprintf("%d", b.Amount),
			Fee:    "0",
			// TODO: fix token
			CrossChainAsset: CrossChainAsset{
				// TODO:
				Name:     "b.Symbol",
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

func (b *BridgeTxsResult) GetDestination() *DestinationDocument {
	return &DestinationDocument{
		BaseDocument: &BaseDocument{
			Chain: b.DestinationChain,
			// TODO:
			ChainName: "b.ChainName",
			TxHash:    b.ExecutedTxHash,

			// TODO: use token approved status or executed status
			Status: b.Status,
			Value:  fmt.Sprintf("%d", b.Amount),
			Fee:    "0",
			CrossChainAsset: CrossChainAsset{
				// TODO:
				Name:     "b.Symbol",
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
