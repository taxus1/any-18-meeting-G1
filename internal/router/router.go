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

	// Item 模块
	svcItem := service.NewItem(repository.NewItem(db), rdb)
	hItem := handler.NewItem(svcItem)

	// 会议室模块
	svcMeeting := service.NewMeetingService(repository.NewRoomRepository(db), repository.NewBookingRepository(db), rdb)
	hRoom := handler.NewRoomHandler(svcMeeting)
	hBooking := handler.NewBookingHandler(svcMeeting)

	api := r.Group("/api")
	{
		// demo 接口
		api.GET("/demo/ping", hItem.Ping)
		api.POST("/demo/item", hItem.CreateItem)
		api.GET("/demo/list", hItem.List)
		api.GET("/demo/cache", hItem.Cache)

		// 会议室接口
		api.POST("/meeting/room", hRoom.CreateRoom)
		api.POST("/meeting/room/disable", hRoom.DisableRoom)
		api.POST("/meeting/booking", hBooking.CreateBooking)
		api.POST("/meeting/booking/cancel", hBooking.CancelBooking)
		api.GET("/meeting/booking/list", hBooking.ListBookings)
	}
	return r
}
