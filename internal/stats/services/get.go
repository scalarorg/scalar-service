package services

import (
	"context"

	"github.com/scalarorg/scalar-service/pkg/db"
)

type StatsOpts struct {
	TimeBucket string `query:"time_bucket" validate:"omitempty,oneof=hour day week month"`
}

type StatsPayload struct {
	Value uint64 `json:"data"`
	Time  int64  `json:"time"`
}

type StatsResponse struct {
	Txs         []*StatsPayload `json:"txs"`
	Volumes     []*StatsPayload `json:"volumes"`
	ActiveUsers []*StatsPayload `json:"active_users"`
	NewUsers    []*StatsPayload `json:"new_users"`
}

func GetStats(ctx context.Context, opts *StatsOpts) (*StatsResponse, error) {
	cmds, err := db.GetCommandStats(ctx, opts.TimeBucket)
	if err != nil {
		return nil, err
	}

	tokenSentSats, err := db.GetTokenStats(opts.TimeBucket)
	if err != nil {
		return nil, err
	}

	txs := make([]*StatsPayload, 0)
	volumes := make([]*StatsPayload, 0)
	activeUsers := make([]*StatsPayload, 0)
	newUsers := make([]*StatsPayload, 0)

	for _, cmd := range cmds {
		txs = append(txs, &StatsPayload{
			Value: cmd.Count,
			Time:  cmd.BucketTime.Unix(),
		})
	}

	for _, token := range tokenSentSats {
		volumes = append(volumes, &StatsPayload{
			Value: token.TotalAmount,
			Time:  token.BucketTime.Unix(),
		})

		activeUsers = append(activeUsers, &StatsPayload{
			Value: token.ActiveUsers,
			Time:  token.BucketTime.Unix(),
		})

		newUsers = append(newUsers, &StatsPayload{
			Value: token.NewUsers,
			Time:  token.BucketTime.Unix(),
		})
	}

	return &StatsResponse{
		Txs:         txs,
		Volumes:     volumes,
		ActiveUsers: activeUsers,
		NewUsers:    newUsers,
	}, nil
}
