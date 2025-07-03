package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/stats/services"
)

func GetTopSourceChainsByTx(c echo.Context) error {
	limitInt := getLimit(c)

	result, err := services.StatTransactionBySourceChain(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopDestinationChainsByTx(c echo.Context) error {
	limitInt := getLimit(c)

	result, err := services.StatTransactionByDestinationChain(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}

func GetTopPathsByTx(c echo.Context) error {
	limitInt := getLimit(c)

	result, err := services.StatTransactionByPath(limitInt)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}
