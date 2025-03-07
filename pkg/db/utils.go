package db

import "github.com/scalarorg/data-models/chains"

func CreateCrossChainDocument(sent *chains.TokenSent) *CrossChainDocument {
	return &CrossChainDocument{
		ID:     "event-id-1",
		Type:   CrossChainTxBridge,
		Status: string(chains.TokenSentStatusSuccess),
		Source: &SourceDocument{
			BaseDocument: &BaseDocument{
				Chain:       "bitcoin|4",
				ChainName:   "Bitcoin Testnet 4",
				TxHash:      "0x1234567890",
				BlockHeight: 1234567890,
				Status:      "success",
				Value:       "0.001",
				Fee:         "0.0001",
				CrossChainAsset: CrossChainAsset{
					Name:     "Bitcoin Testnet 4",
					Symbol:   "tBTC",
					Address:  "0x01",
					Decimals: 8,
					IsNative: true,
				},
				CreatedAt: 1741320000,
			},
			Sender: "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq",
		},
		Destination: &DestinationDocument{
			BaseDocument: &BaseDocument{
				Chain:       "evm|1115511",
				ChainName:   "Bitcoin Testnet 4",
				TxHash:      "0x1234567890",
				BlockHeight: 1234567890,
				Status:      "success",
				Value:       "0.001",
				Fee:         "0.0001",
				CrossChainAsset: CrossChainAsset{
					Name:     "Scalar Bitcoin",
					Symbol:   "sBTC",
					Address:  "0x982321eb5693cdbAadFfe97056BEce07D09Ba49f",
					Decimals: 8,
					IsNative: false,
				},
				CreatedAt: 1741329838,
			},
			Receiver: "0x982321eb5693cdbAadFfe97056BEce07D09Ba49f",
		},
	}
}
