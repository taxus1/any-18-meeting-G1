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

	// 初始化仓储层
	roomRepo := repository.NewRoomRepository(db)
	bookingRepo := repository.NewBookingRepository(db)

	// 初始化服务层
	roomSvc := service.NewRoomService(roomRepo)
	bookingSvc := service.NewBookingService(roomRepo, bookingRepo)

	// 初始化处理器层
	itemHandler := handler.NewItem(service.NewItem(repository.NewItem(db), rdb))
	roomHandler := handler.NewRoomHandler(roomSvc)
	bookingHandler := handler.NewBookingHandler(bookingSvc)

	// demo 接口
	api := r.Group("/api/demo")
	{
		api.GET("/ping", itemHandler.Ping)
		api.POST("/item", itemHandler.CreateItem)
		api.GET("/list", itemHandler.List)
		api.GET("/cache", itemHandler.Cache)
	}

	// 会议室接口
	meeting := r.Group("/api/meeting")
	{
		// 会议室管理
		meeting.POST("/room", roomHandler.CreateRoom)
		meeting.POST("/room/disable", roomHandler.DisableRoom)

		// 预约管理
		meeting.POST("/booking", bookingHandler.Book)
		meeting.POST("/booking/cancel", bookingHandler.Cancel)
		meeting.GET("/booking/list", bookingHandler.List)
	}

	return r
}
