package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/constants"
	"github.com/scalarorg/scalar-service/internal/stats/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type User struct {
	ID string `query:"id"`
}


func GetTxsStatsHandler(c echo.Context) error {
	var opts services.StatsOpts
	if err := utils.BindAndValidate(c, &opts); err != nil {
		return err
	}

	//Set default network to testnet4
	switch opts.Network {
	case "mainnet":
		opts.Network = "bitcoin|1"
	case "testnet":
		opts.Network = constants.DefaultChain
	}

	//Set default limit to 10
	if opts.Limit == 0 {
		opts.Limit = 10
	}	

	txs, err := services.GetTxsStats(c.Request().Context(), &opts)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, txs)
}

func GetVolumesStatsHandler(c echo.Context) error {
	var opts services.StatsOpts
	if err := utils.BindAndValidate(c, &opts); err != nil {
		return err
	}

	//Set default network to testnet4
	switch opts.Network {
	case "mainnet":
		opts.Network = "bitcoin|1"
	case "testnet":
		opts.Network = constants.DefaultChain
	}

	//Set default limit to 10
	if opts.Limit == 0 {
		opts.Limit = 10
	}

	volumes, err := services.GetVolumesStats(c.Request().Context(), &opts)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, volumes)
}

func GetActiveUsersStatsHandler(c echo.Context) error {
	var opts services.StatsOpts
	if err := utils.BindAndValidate(c, &opts); err != nil {
		return err
	}

	//Set default network to testnet4
	switch opts.Network {
	case "mainnet":
		opts.Network = "bitcoin|1"
	case "testnet":
		opts.Network = constants.DefaultChain
	}
	
	//Set default limit to 10
	if opts.Limit == 0 {
		opts.Limit = 10
	}

	activeUsers, err := services.GetActiveUsersStats(c.Request().Context(), &opts)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, activeUsers)
}

func GetNewUsersStatsHandler(c echo.Context) error {
	var opts services.StatsOpts
	if err := utils.BindAndValidate(c, &opts); err != nil {
		return err
	}

	//Set default network to testnet4
	switch opts.Network {
	case "mainnet":
		opts.Network = "bitcoin|1"
	case "testnet":
		opts.Network = constants.DefaultChain
	}

	//Set default limit to 10
	if opts.Limit == 0 {
		opts.Limit = 10
	}

	newUsers, err := services.GetNewUsersStats(c.Request().Context(), &opts)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, newUsers)
}
