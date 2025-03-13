package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/constants"
	"github.com/scalarorg/scalar-service/internal/stats/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

func GetTopUsersByVolume(c echo.Context) error {
	//ctx := c.Request().Context()

	var body services.StatsOpts
	err := utils.BindAndValidate(c, &body)
	if err != nil {
		return err
	}

	limit := c.QueryParam("limit")
	if limit == "" {
		limit = "10"
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.GetTopUsersByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopBridgesByVolume(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = constants.DefaultLimit
	}
	chain := c.QueryParam("chain")
	if chain == "" {
		chain = constants.DefaultChain
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.GetTopBridgesByVolume(chain, limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopSourceChainsByVolume(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = constants.DefaultLimit
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.GetTopSourceChainsByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopDestinationChainsByVolume(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = "10"
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.GetTopDestinationChainsByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopPathsByVolume(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = constants.DefaultLimit
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.GetTopPathsByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}
