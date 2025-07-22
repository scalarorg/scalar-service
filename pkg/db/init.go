package db

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/scalar-service/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBManager struct {
	Relayer *gorm.DB
	Indexer *gorm.DB
}

var DB = &DBManager{}

func Init() {
	DB.Relayer = connect(config.Env.RELAYER_DB_URI)
	DB.Indexer = connect(config.Env.INDEXER_DB_URI)

	// Create optimized indexes for better query performance
	createOptimizedIndexes()
}

func Close() error {
	return DB.Close()
}

func (db *DBManager) Close() error {
	if sqlDB, err := db.Relayer.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close relayer db: %w", err)
		}
	}

	if sqlDB, err := db.Indexer.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close indexer db: %w", err)
		}
	}

	return nil
}

// func runMigrations() {
// 	models := []interface{}{
// 		&chains.MintCommand{},
// 		&chains.CommandExecuted{},
// 		&chains.ContractCall{},
// 		&chains.ContractCallWithToken{},
// 		&chains.TokenSent{},
// 		&scalarnet.TokenSentApproved{},
// 		&scalarnet.Command{},
// 		&scalarnet.BatchCommand{},
// 		&scalarnet.EventCheckPoint{},
// 		&scalarnet.CallContract{},
// 		&scalarnet.CallContractWithToken{},
// 		&scalarnet.ContractCallApproved{},
// 		&scalarnet.ContractCallApprovedWithMint{},
// 	}

// 	if err := DB.Relayer.AutoMigrate(models...); err != nil {
// 		panic(fmt.Sprintf("failed to run migrations: %+v", err))
// 	}

// 	// TODO: Run indexer migrations
// }

func connect(connectionString string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Reduce logging overhead in production
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Optimize GORM settings for better performance
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %+v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get database instance: %+v", err))
	}

	// Optimized connection pool settings
	sqlDB.SetMaxIdleConns(25)                  // Increased from 10
	sqlDB.SetMaxOpenConns(200)                 // Increased from 100
	sqlDB.SetConnMaxLifetime(time.Hour * 2)    // Increased lifetime
	sqlDB.SetConnMaxIdleTime(time.Minute * 30) // Add idle timeout

	return db
}

// createOptimizedIndexes creates database indexes for better query performance
func createOptimizedIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexes := []string{
		// Vault transactions indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_chain ON vault_transactions(chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_destination_chain ON vault_transactions(destination_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_staker_script_pubkey ON vault_transactions(staker_script_pubkey)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_amount ON vault_transactions(amount DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_timestamp ON vault_transactions(timestamp DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_tx_hash ON vault_transactions(tx_hash)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_chain_destination ON vault_transactions(chain, destination_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_composite ON vault_transactions(chain, destination_chain, timestamp)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_vault_transactions_user_time ON vault_transactions(staker_script_pubkey, timestamp)`,

		// Token sents indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_source_chain ON token_sents(source_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_destination_chain ON token_sents(destination_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_source_address ON token_sents(source_address)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_sender ON token_sents(source_address)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_amount ON token_sents(amount DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_block_time ON token_sents(block_time DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_tx_hash ON token_sents(tx_hash)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_token_sents_composite ON token_sents(source_chain, destination_chain, block_time)`,

		// Contract call with tokens indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_source_chain ON contract_call_with_tokens(source_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_destination_chain ON contract_call_with_tokens(destination_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_sender ON contract_call_with_tokens(source_address)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_amount ON contract_call_with_tokens(amount DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_block_number ON contract_call_with_tokens(block_number DESC)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_tx_hash ON contract_call_with_tokens(tx_hash)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_contract_call_with_tokens_composite ON contract_call_with_tokens(source_chain, destination_chain, block_number)`,

		// Command executed indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_command_executeds_command_id ON command_executeds(command_id)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_command_executeds_source_chain ON command_executeds(source_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_command_executeds_block_number ON command_executeds(block_number)`,

		// Block headers indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_block_headers_chain_block_number ON block_headers(chain, block_number)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_block_headers_block_time ON block_headers(block_time DESC)`,

		// EVM redeem txes indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_evm_redeem_txes_destination_chain ON evm_redeem_txes(destination_chain)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_evm_redeem_txes_custodian_group_uid ON evm_redeem_txes(custodian_group_uid)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_evm_redeem_txes_session_sequence ON evm_redeem_txes(session_sequence)`,

		// BTC redeem txes indexes
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_btc_redeem_txes_custodian_group_uid ON btc_redeem_txes(custodian_group_uid)`,
		`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_btc_redeem_txes_session_sequence ON btc_redeem_txes(session_sequence)`,
	}

	// Create indexes on both databases
	for _, indexSQL := range indexes {
		// Try on Indexer DB first
		if err := DB.Indexer.WithContext(ctx).Exec(indexSQL).Error; err != nil {
			log.Warn().Err(err).Str("sql", indexSQL).Msg("Failed to create index on indexer DB")
		}
	}

	log.Info().Msg("Database indexes optimization completed")
}

