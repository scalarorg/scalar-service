package handlers

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/constants"
	"github.com/scalarorg/scalar-service/internal/stats/services"
)

func GetTopSourceChainsByTx(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = constants.DefaultLimit
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.StatTransactionBySourceChain(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopDestinationChainsByTx(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = constants.DefaultLimit
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.StatTransactionByDestinationChain(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopPathsByTx(c echo.Context) error {
	limit := c.QueryParam("limit")
	if limit == "" {
		limit = constants.DefaultLimit
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return err
	}

	result, err := services.StatTransactionByPath(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}
