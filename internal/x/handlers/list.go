package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/x/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

func List(c echo.Context) error {
	ctx := c.Request().Context()

	var body services.ListOptions

	if err := utils.BindAndValidate(c, &body); err != nil {
		return err
	}

	txs, count, err := services.List(ctx, &body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, utils.NewListResult(txs, count))
}
