package router

import (
	"github.com/gin-gonic/gin"
	"github.com/zverobot1k/threat-service-blokhin/internal/handler"
)

func New(threatHandler *handler.ThreatHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.GET("/info/:md5", threatHandler.GetByMD5)
	r.PUT("/info/:md5", threatHandler.Update)
	r.POST("/info", threatHandler.Create)
	return r
}
