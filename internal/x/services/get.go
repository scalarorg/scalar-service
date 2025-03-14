package services

import (
	"context"
	"fmt"

	"github.com/scalarorg/scalar-service/pkg/db"
)

type GetOptions struct {
	TxHash string `param:"tx_hash" validate:"required"`
	Type   string `param:"type" validate:"omitempty,oneof=bridge transfer redeem"`
}

func Get(ctx context.Context, options *GetOptions) (*db.CrossChainDocument, error) {
	var (
		tx  *db.BaseCrossChainTxResult
		err error
	)

	fmt.Printf("options: %+v\n", options)

	if options.Type == "bridge" {
		tx, err = db.GetBridgeTx(ctx, options.TxHash)
	} else if options.Type == "transfer" {
		tx, err = db.GetTransferTx(ctx, options.TxHash)
	} else if options.Type == "redeem" {
		tx, err = db.GetRedeemTx(ctx, options.TxHash)
	}

	if err != nil {
		return nil, err
	}

	result := db.CreateCrossChainDocument(tx)

	return result, nil
}
