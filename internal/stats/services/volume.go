package services

import (
	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/types"
)

func GetTopUsersByVolume(limit int) ([]types.AddressAmount, error) {
	return db.GetTopTransferUsers(limit)
}

func GetTopBridgesByVolume(sourceChain string, limit int) ([]*types.AddressAmount, error) {
	return db.GetTopBridgeUsers(sourceChain, limit)
}

func GetTopSourceChainsByVolume(limit int) ([]*types.ChainAmount, error) {
	return db.StatVolumeBySourceChain(limit)
}

func GetTopDestinationChainsByVolume(limit int) ([]*types.ChainAmount, error) {
	return db.StatVolumeByDestinationChain(limit)
}

func GetTopPathsByVolume(limit int) ([]*types.PathAmount, error) {
	return db.StatVolumeByPath(limit)
}
