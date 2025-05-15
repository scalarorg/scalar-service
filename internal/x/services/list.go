package services

import (
	"context"
	"fmt"

	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type ListOptions struct {
	Size int    `json:"size,omitempty"`
	Page int    `json:"page,omitempty"`
	Type string `json:"type,omitempty" validate:"omitempty,oneof=bridge transfer redeem"`
}

func List(ctx context.Context, options *ListOptions) ([]*db.CrossChainDocument, int, error) {
	var (
		txs   []db.BaseCrossChainTxResult
		count int
		err   error
	)
	//Page starts from 0
	offset := options.Page * options.Size
	if offset < 0 {
		offset = 0
	}

	if options.Type == "bridge" {
		txs, count, err = db.ListBridgeTxs(ctx, options.Size, offset)
	} else if options.Type == "transfer" {
		txs, count, err = db.ListTransferTxs(ctx, options.Size, offset)
	} else if options.Type == "redeem" {
		txs, count, err = db.ListRedeemTxs(ctx, options.Size, offset)
	} else {
		return nil, 0, fmt.Errorf("invalid type")
	}

	if err != nil {
		return nil, 0, err
	}

	list := utils.Map(txs, func(tx db.BaseCrossChainTxResult) *db.CrossChainDocument { return db.CreateCrossChainDocument(&tx) })

	return list, count, nil
}
