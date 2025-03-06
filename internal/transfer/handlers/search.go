package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/transfer/services"
	"github.com/scalarorg/scalar-service/pkg/db"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

func Search(c echo.Context) error {
	ctx := c.Request().Context()

	var body db.Options

	if err := utils.BindAndValidate(c, &body); err != nil {
		return err
	}

	tokenSents, count, err := services.SearchTransfers(ctx, &body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, utils.NewListResult(tokenSents, count))
}
