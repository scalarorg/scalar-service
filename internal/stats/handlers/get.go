package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/constants"
	"github.com/scalarorg/scalar-service/internal/stats/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type User struct {
	ID string `query:"id"`
}

func setDefaultOpts(opts *services.StatsOpts) {
	if opts.Limit == 0 {
		if opts.Size > 0 {
			opts.Limit = opts.Size
		} else {
			opts.Limit = 10
		}
	}
}

func getLimit(c echo.Context) int {
	limit := c.QueryParam("limit")
	size := c.QueryParam("size")
	if limit == "" {
		if size == "" {
			limit = constants.DefaultLimit
		} else {
			limit = size
		}
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return 10
	}
	return limitInt
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
	setDefaultOpts(&opts)

	txs, err := services.GetTxsChartData(c.Request().Context(), &opts)
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
	setDefaultOpts(&opts)

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
	setDefaultOpts(&opts)

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
	setDefaultOpts(&opts)

	newUsers, err := services.GetNewUsersStats(c.Request().Context(), &opts)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, newUsers)
}

func GetSummaryStatsHandler(c echo.Context) error {
	var opts services.StatsOpts
	if err := utils.BindAndValidate(c, &opts); err != nil {
		return err
	}

	switch opts.Network {
	case "mainnet":
		opts.Network = "bitcoin|1"
	case "testnet":
		opts.Network = constants.DefaultChain
	default:
		opts.Network = constants.DefaultChain
	}

	summary, err := services.GetSummaryStats(c.Request().Context(), &opts)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, summary)
}
