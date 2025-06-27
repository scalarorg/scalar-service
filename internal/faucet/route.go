package faucet

import (
	"github.com/labstack/echo/v4"
)

func Route(g *echo.Group, path string) {
	group := g.Group(path)
	group.POST("", handleFaucet)
}
