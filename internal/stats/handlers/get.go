package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/stats/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

type User struct {
	ID string `query:"id"`
}

func Get(c echo.Context) error {
	ctx := c.Request().Context()

	var body services.StatsOpts
	err := utils.BindAndValidate(c, &body)
	if err != nil {
		return err
	}

	result, err := services.GetStats(ctx, &body)
	if err != nil {
		return err
	}

	return c.JSON(200, result)
}
