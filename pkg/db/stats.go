package db

import (
	"context"
	"fmt"
	"time"
)

type Stats struct {
	BucketTime time.Time `json:"bucket_time"`
	Count      uint64    `json:"count"`
}

// mergeSortedEntries merges two sorted slices of Entry
func mergeSortedStats(a, b []Stats) []Stats {
	var result []Stats
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if a[i].BucketTime.Equal(b[j].BucketTime) {
			result = append(result, Stats{
				BucketTime: a[i].BucketTime,
				Count:      a[i].Count + b[j].Count,
			})
			i++
			j++
		} else if a[i].BucketTime.Before(b[j].BucketTime) {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}

	// Append any remaining entries
	for ; i < len(a); i++ {
		result = append(result, a[i])
	}
	for ; j < len(b); j++ {
		result = append(result, b[j])
	}

	return result
}

func GetCommandStats(ctx context.Context, timeBucket string) ([]Stats, error) {
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	var tokenStats []Stats
	err := DB.Relayer.Table("token_sents").
		Select("date_trunc(?, to_timestamp(block_time)) as bucket_time, COUNT(*) as count", timeBucket).
		Group("bucket_time").
		Order("bucket_time ASC").
		Find(&tokenStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch token_sents stats: %w", err)
	}
	var ccwtStats []Stats
	err = DB.Relayer.Table("contract_call_with_tokens ccwt left join block_headers bh on ccwt.source_chain = bh.chain and ccwt.block_number = bh.block_number").
		Select("date_trunc(?, to_timestamp(bh.block_time)) as bucket_time, COUNT(*) as count", timeBucket).
		Group("bucket_time").
		Order("bucket_time ASC").
		Find(&ccwtStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch token_sents stats: %w", err)
	}
	//Merge two list
	allStats := mergeSortedStats(tokenStats, ccwtStats)
	return allStats, nil
}

// func GetCommandStatsWithTimeScale(ctx context.Context, timeBucket string) ([]Stats, error) {
// 	interval := getTimeBucketInterval(timeBucket)

// 	var stats []Stats
// 	err := DB.Relayer.Table("commands").
// 		Select("time_bucket(? :: interval, created_at) as bucket_time, COUNT(*) as count", interval).
// 		Group("bucket_time").
// 		Order("bucket_time ASC").
// 		Find(&stats).Error

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to fetch command stats: %w", err)
// 	}

// 	return stats, nil
// }

type TokenSentStats struct {
	BucketTime  time.Time `json:"bucket_time" gorm:"column:bucket_time"`
	ActiveUsers uint64    `json:"active_users" gorm:"column:active_users"`
	TotalAmount uint64    `json:"total_amount" gorm:"column:total_amount"`
	NewUsers    uint64    `json:"new_users" gorm:"column:new_users"`
}

func GetTokenStats(timeBucket string) ([]TokenSentStats, error) {
	//interval := getTimeBucketInterval(timeBucket)
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	var stats []TokenSentStats
	// timeScaleRawQuery := `
	// 	SELECT
	// 		time_bucket(? :: interval, ts.created_at) as bucket_time,
	// 		COUNT(DISTINCT ts.source_address) as active_users,
	// 		COUNT(DISTINCT CASE WHEN ts.created_at = first_seen.first_time THEN ts.source_address ELSE NULL END) as new_users,
	// 		SUM(ts.amount) as total_amount
	// 	FROM token_sents ts
	// 	JOIN (
	// 		SELECT
	// 			source_address,
	// 			MIN(created_at) as first_time
	// 		FROM token_sents
	// 		GROUP BY source_address
	// 	) as first_seen ON ts.source_address = first_seen.source_address
	// 	GROUP BY bucket_time
	// 	ORDER BY bucket_time ASC
	// `
	rawQuery := `
		SELECT 
			date_trunc(?, to_timestamp(ts.block_time)) as bucket_time,
			COUNT(DISTINCT ts.source_address) as active_users,
			COUNT(DISTINCT ts.source_address) as new_users,
			SUM(ts.amount) as total_amount
		FROM token_sents ts
		GROUP BY bucket_time
		ORDER BY bucket_time ASC 
	`
	err := DB.Relayer.Raw(rawQuery, timeBucket).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	return stats, nil
}

func GetTotalTxs() (int64, error) {
	var totalTxs int64
	query := `
		SELECT
			COUNT(*) as total_txs
		FROM event_token_sents
	`
	err := DB.Indexer.Raw(query).Scan(&totalTxs).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total txs: %w", err)
	}
	return totalTxs, nil
}

func GetTotalBridgedVolumes(chain string) (int64, error) {
	var totalVolumes int64
	query := `
		SELECT
			SUM(asset_amount) as total_volumes
		FROM event_token_sents
		WHERE chain = ?
	`
	err := DB.Indexer.Raw(query, chain).Scan(&totalVolumes).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total volumes: %w", err)
	}
	return totalVolumes, nil
}

func GetTotalUsers() (int64, error) {
	var totalUsers int64
	query := `
		SELECT
			COUNT(DISTINCT sender) as total_users
		FROM event_token_sents
	`
	err := DB.Indexer.Raw(query).Scan(&totalUsers).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total users: %w", err)
	}
	return totalUsers, nil
}
