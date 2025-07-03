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

// Count transactions by time
// Count bridge transaction in vault_transactions for bitcoin tx
// Count contract call with tokens for evm transaction
func GetCommandStats(ctx context.Context, timeBucket string, limit int) ([]Stats, error) {
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	var vaultTxStats []Stats
	go func() {
		defer wg.Done()
		query := `
			SELECT 
				date_trunc(?, to_timestamp(timestamp)) as bucket_time, COUNT(*) as count
			FROM vault_transactions
			GROUP BY bucket_time
			ORDER BY bucket_time DESC
			LIMIT ?
		`
		err := DB.Indexer.Raw(query, timeBucket, limit).Scan(&vaultTxStats).Error
		if err != nil {
			log.Error().Err(err).Msg("failed to fetch vault_transactions stats")
		}
	}()
	var ccwtStats []Stats
	go func() {
		defer wg.Done()
		query := `
			SELECT 
				date_trunc(?, to_timestamp(bh.block_time)) as bucket_time, COUNT(*) as count
			FROM contract_call_with_tokens ccwt 
			LEFT JOIN block_headers bh ON ccwt.source_chain = bh.chain AND ccwt.block_number = bh.block_number
			GROUP BY bucket_time
			ORDER BY bucket_time DESC
			LIMIT ?
		`
		err := DB.Indexer.Raw(query, timeBucket, limit).Scan(&ccwtStats).Error
		if err != nil {
			log.Error().Err(err).Msg("failed to fetch contract_call_with_tokens stats")
		}
	}()
	wg.Wait()
	//Merge two list
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
	return allStats[len(allStats)-limit:], nil
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

func GetStatsByTimeBucket(timeBucket string) ([]TokenSentStats, error) {
	rawQuery := `
	SELECT 
		date_trunc(?, to_timestamp(vt.timestamp)) as bucket_time,
		sum(amount) as total_amount,
		count(distinct staker_script_pubkey) as active_users
		FROM vault_transactions vt
		GROUP BY bucket_time
		order by bucket_time asc`
	var stats []TokenSentStats
	err := DB.Indexer.Raw(rawQuery, timeBucket).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	return stats, nil
}

func GetVolumeByTimeBucket(timeBucket string, limit int) ([]TokenSentStats, error) {
	rawQuery := `
	SELECT 
		date_trunc(?, to_timestamp(vt.timestamp)) as bucket_time,
		sum(amount) as total_amount
		FROM vault_transactions vt
		GROUP BY bucket_time
		order by bucket_time asc
		limit ?`
	var stats []TokenSentStats
	err := DB.Indexer.Raw(rawQuery, timeBucket, limit).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	return stats, nil
}

func GetActiveUsersByTimeBucket(timeBucket string) ([]TokenSentStats, error) {
	rawQuery := `
	SELECT 
		date_trunc(?, to_timestamp(vt.timestamp)) as bucket_time,
		count(distinct staker_script_pubkey) as active_users
		FROM vault_transactions vt
		GROUP BY bucket_time
		order by bucket_time asc`
	var stats []TokenSentStats
	err := DB.Indexer.Raw(rawQuery, timeBucket).Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	return stats, nil
}
func GetNewUsersByTimeBucket(timeBucket string, limit int) ([]TokenSentStats, error) {
	if !validateTimeBucketInterval(timeBucket) {
		return nil, fmt.Errorf("invalid bucket name")
	}
	var stats []TokenSentStats
	rawQuery := `
		select 
			count(staker_script_pubkey) as new_users,
			bucket_time
		from (SELECT
				distinct staker_script_pubkey,
				min(date_trunc(?, to_timestamp(vt.timestamp))) as bucket_time
			FROM vault_transactions vt
			GROUP BY staker_script_pubkey 
		) as first_seen
		group by bucket_time
		order by bucket_time desc
		limit ?
	`
	err := DB.Indexer.Raw(rawQuery, timeBucket, limit).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch token stats: %w", err)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].BucketTime.Before(stats[j].BucketTime)
	})
	return stats, nil
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
			COUNT(DISTINCT CASE WHEN ts.block_time = first_seen.first_time THEN ts.source_address ELSE NULL END) as new_users,
			SUM(ts.amount) as total_amount
		FROM token_sents ts
		JOIN (
			SELECT
				source_address,
				min(block_time) as first_time
			FROM token_sents
			GROUP BY source_address
		) as first_seen ON ts.source_address = first_seen.source_address
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
	// query := `
	// 	SELECT
	// 		COUNT(*) as total_txs
	// 	FROM token_sents
	// `
	// err := DB.Relayer.Raw(query).Scan(&totalTxs).Error
	query := `
		SELECT 
			COUNT(*) as total_txs
		FROM vault_transactions
	`
	err := DB.Indexer.Raw(query).Scan(&totalTxs).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total txs: %w", err)
	}
	return totalTxs, nil
}
func GetTotalBridgedVolumes(chain string) (int64, error) {
	var totalVolumes int64
	// query := `
	// 	SELECT
	// 		SUM(amount) as total_volumes
	// 	FROM token_sents
	// 	WHERE source_chain = ?
	// `
	// err := DB.Relayer.Raw(query, chain).Scan(&totalVolumes).Error
	query := `
		SELECT
			SUM(amount) as total_volumes
		FROM vault_transactions
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
	// query := `
	// 	SELECT
	// 		COUNT(DISTINCT source_address) as total_users
	// 	FROM token_sents
	// `
	// err := DB.Relayer.Raw(query).Scan(&totalUsers).Error
	query := `
		SELECT
			COUNT(DISTINCT staker_script_pubkey) as total_users
		FROM vault_transactions
	`
	err := DB.Indexer.Raw(query).Scan(&totalUsers).Error
	if err != nil {
		return 0, fmt.Errorf("failed to fetch total users: %w", err)
	}
	return totalUsers, nil
}
