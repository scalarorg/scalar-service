package stats

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/stats/handlers"
)

func Route(g *echo.Group, path string) {
	x := g.Group(path)
	volume := x.Group("/volume")
	volume.GET("/topUsers", handlers.GetTopUsersByVolume)
	volume.GET("/topBridges", handlers.GetTopBridgesByVolume)
	volume.GET("/topSourceChains", handlers.GetTopSourceChainsByVolume)
	volume.GET("/topDestinationChains", handlers.GetTopDestinationChainsByVolume)
	volume.GET("/topPaths", handlers.GetTopPathsByVolume)
	transaction := x.Group("/transaction")
	transaction.GET("/topSourceChains", handlers.GetTopSourceChainsByTx)
	transaction.GET("/topDestinationChains", handlers.GetTopDestinationChainsByTx)
	transaction.GET("/topPaths", handlers.GetTopPathsByTx)
	x.GET("", handlers.Get)
}
