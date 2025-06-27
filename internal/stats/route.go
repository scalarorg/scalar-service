package stats

import (
	"github.com/labstack/echo/v4"
	"github.com/scalarorg/scalar-service/internal/stats/handlers"
)

func Route(g *echo.Group, path string) {
	x := g.Group(path)
	volume := x.Group("/volume")
	volume.GET("/top-users", handlers.GetTopUsersByVolume)
	volume.GET("/top-bridges", handlers.GetTopBridgesByVolume)
	volume.GET("/top-source-chains", handlers.GetTopSourceChainsByVolume)
	volume.GET("/top-destination-chains", handlers.GetTopDestinationChainsByVolume)
	volume.GET("/top-paths", handlers.GetTopPathsByVolume)

	transaction := x.Group("/transaction")
	transaction.GET("/top-source-chains", handlers.GetTopSourceChainsByTx)
	transaction.GET("/top-destination-chains", handlers.GetTopDestinationChainsByTx)
	transaction.GET("/top-paths", handlers.GetTopPathsByTx)

	chart := x.Group("/chart")
	chart.GET("/txs", handlers.GetTxsStatsHandler)
	chart.GET("/volumes", handlers.GetVolumesStatsHandler)
	chart.GET("/active-users", handlers.GetActiveUsersStatsHandler)
	chart.GET("/new-users", handlers.GetNewUsersStatsHandler)

	x.GET("/summary", handlers.GetSummaryStatsHandler)
}
