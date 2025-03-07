package services

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/data-models/chains"
	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type ListOptions struct {
	Size   int    `json:"size,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Type   string `json:"type,omitempty" validate:"omitempty,oneof=bridge transfer redeem"`
}

func List(ctx context.Context, options *ListOptions) ([]*db.CrossChainDocument, int, error) {
	txs, count, err := db.ListCrossChainTxs[chains.TokenSent](ctx, options.Size, options.Offset)
	if err != nil {
		return nil, 0, err
	}

	log.Info().Msgf("Found %d cross chain transactions", count)

	list := utils.Map(txs, func(tx chains.TokenSent) *db.CrossChainDocument { return db.CreateCrossChainDocument(&tx) })
	return list, count, nil
}
