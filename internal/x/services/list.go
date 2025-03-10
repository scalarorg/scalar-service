package services

import (
	"context"

	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type ListOptions struct {
	Size   int    `json:"size,omitempty"`
	Offset int    `json:"offset,omitempty"`
	Type   string `json:"type,omitempty" validate:"omitempty,oneof=bridge transfer redeem"`
}

func List(ctx context.Context, options *ListOptions) ([]*db.CrossChainDocument, int, error) {
	if options.Type == "bridge" {
		bridgeTxs, count, err := db.ListBridgeTxs(ctx, options.Size, options.Offset)
		if err != nil {
			return nil, 0, err
		}

		list := utils.Map(bridgeTxs, func(tx db.BridgeTxsResult) *db.CrossChainDocument { return db.CreateCrossChainDocument(&tx) })
		return list, count, nil
	}

	return nil, 0, nil

}
