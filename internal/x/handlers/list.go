package handlers

import (
	"net/http"
	"strconv"

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

func ListWithQuery(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse query parameters
	var options services.ListOptions
	
	// Set default values
	options.Size = 10
	options.Page = 0
	
	// Parse size parameter
	if sizeStr := c.QueryParam("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			options.Size = size
		}
	}
	
	// Parse page parameter
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page >= 0 {
			options.Page = page
		}
	}
	
	// Parse type parameter
	if typeParam := c.QueryParam("type"); typeParam != "" {
		options.Type = typeParam
	}

	txs, count, err := services.List(ctx, &options)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, utils.NewListResult(txs, count))
}
