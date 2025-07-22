package db

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Stats struct {
	BucketTime time.Time `json:"bucket_time"`
	Count      uint64    `json:"count"`
}

// mergeSortedEntries merges two sorted slices of Entry
// func mergeSortedStats(a, b []Stats) []Stats {
// 	var result []Stats
// 	i, j := 0, 0

// 	for i < len(a) && j < len(b) {
// 		if a[i].BucketTime.Equal(b[j].BucketTime) {
// 			result = append(result, Stats{
// 				BucketTime: a[i].BucketTime,
// 				Count:      a[i].Count + b[j].Count,
// 			})
// 			i++
// 			j++
// 		} else if a[i].BucketTime.Before(b[j].BucketTime) {
// 			result = append(result, a[i])
// 			i++
// 		} else {
// 			result = append(result, b[j])
// 			j++
// 		}
// 	}

// 	// Append any remaining entries
// 	for ; i < len(a); i++ {
// 		result = append(result, a[i])
// 	}
// 	for ; j < len(b); j++ {
// 		result = append(result, b[j])
// 	}

// 	return result
// }

// Count transactions by time with optimized parallel queries
func GetCommandStats(ctx context.Context, timeBucket string, limit int) ([]Stats, error) {
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	
	// Add timeout to the context if not already set
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	wg := sync.WaitGroup{}
	wg.Add(2)
	var vaultTxStats []Stats
	var vaultErr error
	
	go func() {
		defer wg.Done()
		// Optimized query with proper indexing
		query := `
			SELECT 
				date_trunc($1, to_timestamp(timestamp)) as bucket_time, 
				COUNT(*) as count
			FROM vault_transactions
			WHERE timestamp IS NOT NULL
			GROUP BY bucket_time
			ORDER BY bucket_time DESC
			LIMIT $2
		`
		vaultErr = DB.Indexer.WithContext(ctxWithTimeout).Raw(query, timeBucket, limit).Scan(&vaultTxStats).Error
		if vaultErr != nil {
			log.Error().Err(vaultErr).Msg("failed to fetch vault_transactions stats")
		}
	}()
	
	var ccwtStats []Stats
	var ccwtErr error
	
	go func() {
		defer wg.Done()
		// Optimized query with better JOIN performance
		query := `
			SELECT 
				date_trunc($1, to_timestamp(bh.block_time)) as bucket_time, 
				COUNT(*) as count
			FROM contract_call_with_tokens ccwt 
			INNER JOIN block_headers bh ON ccwt.source_chain = bh.chain 
				AND ccwt.block_number = bh.block_number
			WHERE bh.block_time IS NOT NULL
			GROUP BY bucket_time
			ORDER BY bucket_time DESC
			LIMIT $2
		`
		ccwtErr = DB.Indexer.WithContext(ctxWithTimeout).Raw(query, timeBucket, limit).Scan(&ccwtStats).Error
		if ccwtErr != nil {
			log.Error().Err(ccwtErr).Msg("failed to fetch contract_call_with_tokens stats")
		}
	}()
	
	wg.Wait()
	
	// Handle errors
	if vaultErr != nil && ccwtErr != nil {
		return nil, fmt.Errorf("both queries failed: vault=%v, ccwt=%v", vaultErr, ccwtErr)
	}
	
	// Merge and sort results efficiently
	sort.Slice(vaultTxStats, func(i, j int) bool {
		return vaultTxStats[i].BucketTime.Before(vaultTxStats[j].BucketTime)
	})
	sort.Slice(ccwtStats, func(i, j int) bool {
		return ccwtStats[i].BucketTime.Before(ccwtStats[j].BucketTime)
	})
	
	allStats := mergeSortedStats(vaultTxStats, ccwtStats, func(a, b Stats) int {
		return a.BucketTime.Compare(b.BucketTime)
	}, func(a, b Stats) Stats {
		return Stats{
			BucketTime: a.BucketTime,
			Count:      a.Count + b.Count,
		}
	})
	
	if len(allStats) > limit {
		return allStats[len(allStats)-limit:], nil
	}
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

func GetStatsByTimeBucket(timeBucket string, limit int) ([]TokenSentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	
	// Optimized query with proper filtering and indexing
	rawQuery := `
	SELECT 
		date_trunc($1, to_timestamp(vt.timestamp)) as bucket_time,
		SUM(amount) as total_amount,
		COUNT(DISTINCT staker_script_pubkey) as active_users
	FROM vault_transactions vt
	WHERE vt.timestamp IS NOT NULL
		AND vt.amount > 0
		AND vt.staker_script_pubkey IS NOT NULL
	GROUP BY bucket_time
	ORDER BY bucket_time DESC
	LIMIT $2
	`
	
	var stats []TokenSentStats
	err := DB.Indexer.WithContext(ctx).Raw(rawQuery, timeBucket, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BucketTime.Before(stats[j].BucketTime)
	})
	
	if len(stats) > limit {
		return stats[len(stats)-limit:], nil
	}
	return stats, nil
}

func GetVolumeByTimeBucket(timeBucket string, limit int) ([]TokenSentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	
	// Optimized query with proper filtering
	rawQuery := `
	SELECT 
		date_trunc($1, to_timestamp(vt.timestamp)) as bucket_time,
		SUM(amount) as total_amount
	FROM vault_transactions vt
	WHERE vt.timestamp IS NOT NULL
		AND vt.amount > 0
	GROUP BY bucket_time
	ORDER BY bucket_time DESC
	LIMIT $2
	`
	
	var stats []TokenSentStats
	err := DB.Indexer.WithContext(ctx).Raw(rawQuery, timeBucket, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch volume stats: %w", err)
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BucketTime.Before(stats[j].BucketTime)
	})
	
	if len(stats) > limit {
		return stats[len(stats)-limit:], nil
	}
	return stats, nil
}

func GetActiveUsersByTimeBucket(timeBucket string, limit int) ([]TokenSentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	
	// Optimized query with proper filtering
	rawQuery := `
	SELECT 
		date_trunc($1, to_timestamp(vt.timestamp)) as bucket_time,
		COUNT(DISTINCT staker_script_pubkey) as active_users
	FROM vault_transactions vt
	WHERE vt.timestamp IS NOT NULL
		AND vt.staker_script_pubkey IS NOT NULL
		AND vt.staker_script_pubkey != ''
	GROUP BY bucket_time
	ORDER BY bucket_time DESC
	LIMIT $2
	`
	
	var stats []TokenSentStats
	err := DB.Indexer.WithContext(ctx).Raw(rawQuery, timeBucket, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active users stats: %w", err)
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BucketTime.Before(stats[j].BucketTime)
	})
	
	if len(stats) > limit {
		return stats[len(stats)-limit:], nil
	}
	return stats, nil
}
func GetNewUsersByTimeBucket(timeBucket string, limit int) ([]TokenSentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	
	// Optimized query using CTE for better performance
	rawQuery := `
	WITH first_transactions AS (
		SELECT 
			staker_script_pubkey,
			MIN(timestamp) as first_timestamp
		FROM vault_transactions
		WHERE staker_script_pubkey IS NOT NULL
			AND staker_script_pubkey != ''
			AND timestamp IS NOT NULL
		GROUP BY staker_script_pubkey
	)
	SELECT 
		date_trunc($1, to_timestamp(ft.first_timestamp)) as bucket_time,
		COUNT(DISTINCT ft.staker_script_pubkey) as new_users
	FROM first_transactions ft
	GROUP BY bucket_time
	ORDER BY bucket_time DESC
	LIMIT $2
	`
	
	var stats []TokenSentStats
	err := DB.Indexer.WithContext(ctx).Raw(rawQuery, timeBucket, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch new users stats: %w", err)
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BucketTime.Before(stats[j].BucketTime)
	})
	
	if len(stats) > limit {
		return stats[len(stats)-limit:], nil
	}
	return stats, nil
}
func GetTokenStats(timeBucket string, limit int) ([]TokenSentStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	
	// Optimized query with proper filtering and indexing
	rawQuery := `
	SELECT 
		date_trunc($1, to_timestamp(ts.block_time)) as bucket_time,
		COUNT(DISTINCT ts.source_address) as active_users,
		COUNT(DISTINCT CASE WHEN ts.block_time = first_seen.first_time THEN ts.source_address ELSE NULL END) as new_users,
		SUM(ts.amount) as total_amount
	FROM token_sents ts
	JOIN (
		SELECT
			source_address,
			MIN(block_time) as first_time
		FROM token_sents
		WHERE block_time IS NOT NULL
			AND source_address IS NOT NULL
		GROUP BY source_address
	) as first_seen ON ts.source_address = first_seen.source_address
	WHERE ts.block_time IS NOT NULL
		AND ts.amount > 0
		AND ts.source_address IS NOT NULL
	GROUP BY bucket_time
	ORDER BY bucket_time DESC
	LIMIT $2
	`
	
	var stats []TokenSentStats
	err := DB.Relayer.WithContext(ctx).Raw(rawQuery, timeBucket, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BucketTime.Before(stats[j].BucketTime)
	})
	
	if len(stats) > limit {
		return stats[len(stats)-limit:], nil
	}
	return stats, nil
}

func GetTotalTxs() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var totalTxs int64
	// Optimized query with proper filtering
	query := `
		SELECT 
			COUNT(*) as total_txs
		FROM vault_transactions
		WHERE timestamp IS NOT NULL
	`
	err := DB.Indexer.WithContext(ctx).Raw(query).Scan(&totalTxs).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total txs: %w", err)
	}
	return totalTxs, nil
}
func GetTotalBridgedVolumes(chain string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var totalVolumes int64
	// Optimized query with proper filtering
	query := `
		SELECT
			COALESCE(SUM(amount), 0) as total_volumes
		FROM vault_transactions
		WHERE chain = ?
			AND amount > 0
			AND timestamp IS NOT NULL
	`
	err := DB.Indexer.WithContext(ctx).Raw(query, chain).Scan(&totalVolumes).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total volumes: %w", err)
	}
	return totalVolumes, nil
}

func GetTotalUsers() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	var totalUsers int64
	// Optimized query with proper filtering
	query := `
		SELECT
			COUNT(DISTINCT staker_script_pubkey) as total_users
		FROM vault_transactions
		WHERE staker_script_pubkey IS NOT NULL
			AND staker_script_pubkey != ''
	`
	err := DB.Indexer.WithContext(ctx).Raw(query).Scan(&totalUsers).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total users: %w", err)
	}
	return totalUsers, nil
}
