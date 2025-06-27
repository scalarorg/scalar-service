package services

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/types"
)

type StatsOpts struct {
	Limit      int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Network    string `query:"network" validate:"omitempty,oneof=mainnet testnet"`
	TimeBucket string `query:"time_bucket" validate:"omitempty,oneof=hour day week month"`
}

type StatsPayload struct {
	Value uint64 `json:"data"`
	Time  int64  `json:"time"`
}

type StatsResponse struct {
	TotalTxs                     int64                  `json:"total_txs"`
	TotalVolumes                 int64                  `json:"total_volumes"`
	TotalUsers                   int64                  `json:"total_users"`
	Txs                          []*StatsPayload        `json:"txs"`
	Volumes                      []*StatsPayload        `json:"volumes"`
	ActiveUsers                  []*StatsPayload        `json:"active_users"`
	NewUsers                     []*StatsPayload        `json:"new_users"`
	TopUsers                     []types.AddressAmount  `json:"top_users"`
	TopBridges                   []*types.AddressAmount `json:"top_bridges"`
	TopSourceChainsByVolume      []*types.ChainAmount   `json:"top_source_chains_by_volume"`
	TopDestinationChainsByVolume []*types.ChainAmount   `json:"top_destination_chains_by_volume"`
	TopPathsByVolume             []*types.PathAmount    `json:"top_paths_by_volume"`
	TopSourceChainsByTx          []*types.ChainAmount   `json:"top_source_chains_by_tx"`
	TopDestinationChainsByTx     []*types.ChainAmount   `json:"top_destination_chains_by_tx"`
	TopPathsByTx                 []*types.PathAmount    `json:"top_paths_by_tx"`
}

// Todo: Consider using graphql for seperate stat request
func GetStats(ctx context.Context, opts *StatsOpts) (*StatsResponse, error) {
	cmds, err := db.GetCommandStats(ctx, opts.TimeBucket)
	if err != nil {
		return nil, err
	}
	response := &StatsResponse{}
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

	response.Txs = txs
	response.Volumes = volumes
	response.ActiveUsers = activeUsers
	response.NewUsers = newUsers
	response = GetOverallStats(ctx, opts, response)
	response = GetVolumeStats(ctx, opts, response)
	response = GetTransactionStats(ctx, opts, response)
	return response, nil
}

func GetOverallStats(ctx context.Context, opts *StatsOpts, response *StatsResponse) *StatsResponse {
	totalTxs, err := db.GetTotalTxs()
	if err != nil {
		log.Error().Err(err).Msg("failed to get total txs")
	}
	response.TotalTxs = totalTxs
	response.TotalVolumes, err = db.GetTotalBridgedVolumes(opts.Network)
	if err != nil {
		log.Error().Err(err).Msg("failed to get total volumes")
	}
	response.TotalUsers, err = db.GetTotalUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to get total users")
	}
	return response
}

type SummaryStats struct {
	TotalTxs     int64 `json:"total_txs"`
	TotalVolumes int64 `json:"total_volumes"`
	TotalUsers   int64 `json:"total_users"`
}

func GetSummaryStats(ctx context.Context, opts *StatsOpts) (*SummaryStats, error) {
	totalTxs, err := db.GetTotalTxs()
	if err != nil {
		return nil, err
	}
	totalVolumes, err := db.GetTotalBridgedVolumes(opts.Network)
	if err != nil {
		return nil, err
	}
	totalUsers, err := db.GetTotalUsers()
	if err != nil {
		return nil, err
	}
	return &SummaryStats{
		TotalTxs:     totalTxs,
		TotalVolumes: totalVolumes,
		TotalUsers:   totalUsers,
	}, nil
}

func GetVolumeStats(ctx context.Context, opts *StatsOpts, response *StatsResponse) *StatsResponse {
	var err error
	response.TopUsers, err = db.GetTopTransferUsers(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top transfer users")
	}
	response.TopBridges, err = db.GetTopBridgeUsers(opts.Network, opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top bridge users")
	}
	response.TopSourceChainsByVolume, err = db.StatVolumeBySourceChain(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top source chains by volume")
	}
	response.TopDestinationChainsByVolume, err = db.StatVolumeByDestinationChain(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top destination chains by volume")
	}
	response.TopPathsByVolume, err = db.StatVolumeByPath(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top paths by volume")
	}
	return response
}

func GetTransactionStats(ctx context.Context, opts *StatsOpts, response *StatsResponse) *StatsResponse {
	var err error
	response.TopSourceChainsByTx, err = db.StatTransactionBySourceChain(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top source chains by tx")
	}
	response.TopDestinationChainsByTx, err = db.StatTransactionByDestinationChain(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top destination chains by tx")
	}
	response.TopPathsByTx, err = db.StatTransactionByPath(opts.Limit)
	if err != nil {
		log.Error().Err(err).Msg("failed to get top paths by tx")
	}
	return response
}

func GetTxsStats(ctx context.Context, opts *StatsOpts) ([]*StatsPayload, error) {
	cmds, err := db.GetCommandStats(ctx, opts.TimeBucket)
	if err != nil {
		return nil, err
	}
	txs := make([]*StatsPayload, 0)
	for _, cmd := range cmds {
		txs = append(txs, &StatsPayload{
			Value: cmd.Count,
			Time:  cmd.BucketTime.Unix(),
		})
	}
	return txs, nil
}

func GetVolumesStats(ctx context.Context, opts *StatsOpts) ([]*StatsPayload, error) {
	tokenSentSats, err := db.GetTokenStats(opts.TimeBucket)
	if err != nil {
		return nil, err
	}
	volumes := make([]*StatsPayload, 0)
	for _, token := range tokenSentSats {
		volumes = append(volumes, &StatsPayload{
			Value: token.TotalAmount,
			Time:  token.BucketTime.Unix(),
		})
	}
	return volumes, nil
}

func GetActiveUsersStats(ctx context.Context, opts *StatsOpts) ([]*StatsPayload, error) {
	tokenSentSats, err := db.GetTokenStats(opts.TimeBucket)
	if err != nil {
		return nil, err
	}
	activeUsers := make([]*StatsPayload, 0)
	for _, token := range tokenSentSats {
		activeUsers = append(activeUsers, &StatsPayload{
			Value: token.ActiveUsers,
			Time:  token.BucketTime.Unix(),
		})
	}
	return activeUsers, nil
}

func GetNewUsersStats(ctx context.Context, opts *StatsOpts) ([]*StatsPayload, error) {
	tokenSentSats, err := db.GetTokenStats(opts.TimeBucket)
	if err != nil {
		return nil, err
	}
	newUsers := make([]*StatsPayload, 0)
	for _, token := range tokenSentSats {
		newUsers = append(newUsers, &StatsPayload{
			Value: token.NewUsers,
			Time:  token.BucketTime.Unix(),
		})
	}
	return newUsers, nil
}
