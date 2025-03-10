package db

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
