package router

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"somepro/internal/handler"
	"somepro/internal/middleware"
	"somepro/internal/repository"
	"somepro/internal/service"
)

// New 组装路由。验收测试契约无关：运行时用 gin 的 Routes() 扫路由，不写死类名/方法名。
func New(db *gorm.DB, rdb *redis.Client) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middleware.OperatorAuth())

	svc := service.NewItem(repository.NewItem(db), rdb)
	h := handler.NewItem(svc)

	api := r.Group("/api/demo")
	{
		api.GET("/ping", h.Ping)
		api.POST("/item", h.CreateItem)
		api.GET("/list", h.List)
		api.GET("/cache", h.Cache)
	}
	return r
}
