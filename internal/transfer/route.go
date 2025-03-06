package transfer

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/transfer/handlers"
)

func Route(g *echo.Group, path string) {
	transferGr := g.Group(path)

	transferGr.POST("/search", handlers.Search)
}
