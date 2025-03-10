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

var DB *gorm.DB

func Init() {
	db, err := gorm.Open(postgres.Open(config.Env.POSTGRES_URI), &gorm.Config{
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

	DB = db

	runMigrations()
}

func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get database instance: %+v", err))
	}

	if err := sqlDB.Close(); err != nil {
		panic(fmt.Sprintf("failed to close database: %+v", err))
	}
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

	if err := DB.AutoMigrate(models...); err != nil {
		panic(fmt.Sprintf("failed to run migrations: %+v", err))
	}
}
