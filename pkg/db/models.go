package db

import (
	"fmt"
	"time"

	"github.com/scalarorg/bitcoin-vault/go-utils/chain"
	"github.com/scalarorg/data-models/chains"
	"github.com/scalarorg/scalar-service/config"
)

type ExpectedCrossChainDocument interface {
	GetID() string
	GetType() CrossChainTx
	GetStatus() string
	GetSource() *SourceDocument
	GetDestination() *DestinationDocument
	GetCommandID() string
}

type CrossChainTx string

const (
	CrossChainTxBridge   CrossChainTx = "bridge"
	CrossChainTxTransfer CrossChainTx = "transfer"
	CrossChainTxRedeem   CrossChainTx = "redeem"
)

type CrossChainDocument struct {
	ID          string               `json:"id"`
	Type        CrossChainTx         `json:"type"`
	Status      string               `json:"status"`
	CommandID   string               `json:"command_id"`
	Source      *SourceDocument      `json:"source"`
	Destination *DestinationDocument `json:"destination"`
}

type CrossChainAsset struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Decimals uint8  `json:"decimals"`
	IsNative bool   `json:"is_native"`
}

type BaseDocument struct {
	Chain       string `json:"chain"`
	ChainName   string `json:"chain_name"`
	TxHash      string `json:"tx_hash"`
	BlockHeight uint64 `json:"block_height"`
	Status      string `json:"status"`

	Value           string          `json:"value"`
	Fee             string          `json:"fee"`
	CrossChainAsset CrossChainAsset `json:"asset"`
	CreatedAt       uint64          `json:"created_at"`
}

type SourceDocument struct {
	*BaseDocument
	Sender string `json:"sender"`
}

type DestinationDocument struct {
	*BaseDocument
	Receiver string `json:"receiver"`
}

type BaseCrossChainTxResult struct {
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

	//CreatedAt                time.Time `gorm:"column:created_at"`
	CreatedAt                time.Time `gorm:"column:source_created_at"`
	ExecutedCommandCreatedAt time.Time `gorm:"column:executed_created_at"`
}

var _ (ExpectedCrossChainDocument) = (*BaseCrossChainTxResult)(nil)

func (b *BaseCrossChainTxResult) GetID() string {
	return b.EventID
}

func (b *BaseCrossChainTxResult) GetType() CrossChainTx {
	typ := CrossChainTxBridge
	if b.SourceChain != config.Env.BITCOIN_CHAIN_ID {
		typ = CrossChainTxTransfer
	}
	if b.DestinationChain == config.Env.BITCOIN_CHAIN_ID {
		typ = CrossChainTxRedeem
	}
	return typ
}

func (b *BaseCrossChainTxResult) GetStatus() string {
	return b.Status
}

func (b *BaseCrossChainTxResult) GetSource() *SourceDocument {
	var c = &chain.ChainInfo{}
	var name = ""
	err := c.FromString(b.SourceChain)
	if err != nil {
		name = b.SourceChain
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

func (b *BaseCrossChainTxResult) GetDestination() *DestinationDocument {
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

func (b *BaseCrossChainTxResult) GetCommandID() string {
	return b.CommandID
}
