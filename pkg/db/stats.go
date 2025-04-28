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

func GetCommandStats(ctx context.Context, timeBucket string) ([]Stats, error) {
	interval := getTimeBucketInterval(timeBucket)

	var stats []Stats
	err := DB.Relayer.Table("commands").
		Select("time_bucket(? :: interval, created_at) as bucket_time, COUNT(*) as count", interval).
		Group("bucket_time").
		Order("bucket_time ASC").
		Find(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch command stats: %w", err)
	}

	return stats, nil
}

type TokenSentStats struct {
	BucketTime  time.Time `json:"bucket_time" gorm:"column:bucket_time"`
	ActiveUsers uint64    `json:"active_users" gorm:"column:active_users"`
	TotalAmount uint64    `json:"total_amount" gorm:"column:total_amount"`
	NewUsers    uint64    `json:"new_users" gorm:"column:new_users"`
}

func GetTokenStats(timeBucket string) ([]TokenSentStats, error) {
	interval := getTimeBucketInterval(timeBucket)
	var stats []TokenSentStats

	err := DB.Relayer.Raw(`
		SELECT 
			time_bucket(? :: interval, ts.created_at) as bucket_time,
			COUNT(DISTINCT ts.source_address) as active_users,
			COUNT(DISTINCT CASE WHEN ts.created_at = first_seen.first_time THEN ts.source_address ELSE NULL END) as new_users,
			SUM(ts.amount) as total_amount
		FROM token_sents ts
		JOIN (
			SELECT 
				source_address, 
				MIN(created_at) as first_time
			FROM token_sents
			GROUP BY source_address
		) as first_seen ON ts.source_address = first_seen.source_address
		GROUP BY bucket_time
		ORDER BY bucket_time ASC 
	`, interval).Scan(&stats).Error

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
