package services

import (
	"context"
	"fmt"

	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type ListOptions struct {
	Size   int    `json:"size,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Type   string `json:"type,omitempty" validate:"omitempty,oneof=bridge transfer redeem"`
}

func List(ctx context.Context, options *ListOptions) ([]*db.CrossChainDocument, int, error) {
	var (
		txs   []db.BaseCrossChainTxResult
		count int
		err   error
	)

	if options.Type == "bridge" {
		txs, count, err = db.ListBridgeTxs(ctx, options.Size, options.Offset)
	} else if options.Type == "transfer" {
		txs, count, err = db.ListTransferTxs(ctx, options.Size, options.Offset)
	} else if options.Type == "redeem" {
		txs, count, err = db.ListRedeemTxs(ctx, options.Size, options.Offset)
	} else {
		return nil, 0, fmt.Errorf("invalid type")
	}

	if err != nil {
		return nil, 0, err
	}

	list := utils.Map(txs, func(tx db.BaseCrossChainTxResult) *db.CrossChainDocument { return db.CreateCrossChainDocument(&tx) })

	return list, count, nil
}
