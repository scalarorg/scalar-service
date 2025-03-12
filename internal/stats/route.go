package stats

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/stats/handlers"
)

func Route(g *echo.Group, path string) {
	x := g.Group(path)

	x.GET("", handlers.Get)
}
