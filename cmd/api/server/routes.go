package server

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/health"
	"github.com/scalarorg/scalar-service/internal/x"
)

func setupRoute(e *echo.Echo) {
	api := e.Group("/api")
	health.Route(e, "/health")
	x.Route(api, "/x")
}
