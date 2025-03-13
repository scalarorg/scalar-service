package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/constants"
	"github.com/scalarorg/scalar-service/internal/stats/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type User struct {
	ID string `query:"id"`
}

func Get(c echo.Context) error {
	ctx := c.Request().Context()

	var params services.StatsOpts
	err := utils.BindAndValidate(c, &params)
	if err != nil {
		return err
	}
	//Set default network to testnet4
	switch params.Network {
	case "mainnet":
		params.Network = "bitcoin|1"
	case "testnet":
		params.Network = constants.DefaultChain
	default:
		params.Network = constants.DefaultChain
	}
	//Set default limit to 10
	if params.Limit == 0 {
		params.Limit = 10
	}
	result, err := services.GetStats(ctx, &params)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}
