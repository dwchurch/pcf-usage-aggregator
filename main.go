package main

import (
	"pcf-usage-aggregator/metrics"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	v1 := r.Group("/v1")
	{
		v1.GET("/apps", metrics.GetAppData)
	}
	r.Run(":8080")
}
