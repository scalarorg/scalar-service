package db

import (
	"fmt"
	"time"

	"github.com/scalarorg/data-models/chains"
	"github.com/scalarorg/data-models/scalarnet"
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
	runMigrations()
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

func runMigrations() {
	models := []interface{}{
		&chains.MintCommand{},
		&chains.CommandExecuted{},
		&chains.ContractCall{},
		&chains.ContractCallWithToken{},
		&chains.TokenSent{},
		&scalarnet.TokenSentApproved{},
		&scalarnet.Command{},
		&scalarnet.BatchCommand{},
		&scalarnet.EventCheckPoint{},
		&scalarnet.CallContract{},
		&scalarnet.CallContractWithToken{},
		&scalarnet.ContractCallApproved{},
		&scalarnet.ContractCallApprovedWithMint{},
	}

	if err := DB.Relayer.AutoMigrate(models...); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %+v", err))
	}

	// TODO: Run indexer migrations
}

func connect(connectionString string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %+v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get database instance: %+v", err))
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db
}
