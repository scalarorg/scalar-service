package handlers

import (
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

	limitInt := getLimit(c)

	result, err := services.GetTopUsersByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopBridgesByVolume(c echo.Context) error {
	limitInt := getLimit(c)
	chain := c.QueryParam("chain")
	if chain == "" {
		chain = constants.DefaultChain
	}

	result, err := services.GetTopBridgesByVolume(chain, limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopSourceChainsByVolume(c echo.Context) error {
	limitInt := getLimit(c)

	result, err := services.GetTopSourceChainsByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopDestinationChainsByVolume(c echo.Context) error {
	limitInt := getLimit(c)

	result, err := services.GetTopDestinationChainsByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopPathsByVolume(c echo.Context) error {
	limitInt := getLimit(c)

	result, err := services.GetTopPathsByVolume(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}
