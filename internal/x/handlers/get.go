package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/x/services"
	"github.com/scalarorg/scalar-service/pkg/utils"
)

func Get(c echo.Context) error {
	ctx := c.Request().Context()

	var req services.GetOptions

	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}

	tx, err := services.Get(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, tx)
}
