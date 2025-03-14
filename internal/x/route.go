package x

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/x/handlers"
)

func Route(g *echo.Group, path string) {
	x := g.Group(path)

	x.POST("", handlers.List)
	x.GET("/:type/:tx_hash", handlers.Get)
}
