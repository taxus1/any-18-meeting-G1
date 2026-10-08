package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"somepro/internal/common"
	"somepro/internal/service"
)

// RoomHandler 会议室处理器
type RoomHandler struct {
	roomSvc *service.RoomService
}

func NewRoomHandler(roomSvc *service.RoomService) *RoomHandler {
	return &RoomHandler{roomSvc: roomSvc}
}

// CreateRoom POST /api/meeting/room
// 登记会议室
// 参数：roomNo, name, capacity
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	roomNo := c.PostForm("roomNo")
	name := c.PostForm("name")
	capacityStr := c.PostForm("capacity")

	if roomNo == "" || name == "" {
		c.JSON(http.StatusOK, common.Fail("roomNo 和 name 不能为空"))
		return
	}

	capacity := 0
	if capacityStr != "" {
		if n, err := strconv.Atoi(capacityStr); err == nil {
			capacity = n
		}
	}

	room, err := h.roomSvc.CreateRoom(c.Request.Context(), roomNo, name, capacity)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(room))
}

// DisableRoom POST /api/meeting/room/disable
// 停用会议室
// 参数：roomNo
func (h *RoomHandler) DisableRoom(c *gin.Context) {
	roomNo := c.PostForm("roomNo")
	if roomNo == "" {
		c.JSON(http.StatusOK, common.Fail("roomNo 不能为空"))
		return
	}

	if err := h.roomSvc.DisableRoom(c.Request.Context(), roomNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(nil))
}

// BookingHandler 预约处理器
type BookingHandler struct {
	bookingSvc *service.BookingService
}

func NewBookingHandler(bookingSvc *service.BookingService) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc}
}

// Book POST /api/meeting/booking
// 预约会议室
// 参数：roomNo, booker, start, end (时间格式: "2026-01-01 10:00:00")
func (h *BookingHandler) Book(c *gin.Context) {
	roomNo := c.PostForm("roomNo")
	booker := c.PostForm("booker")
	startStr := c.PostForm("start")
	endStr := c.PostForm("end")

	if roomNo == "" || booker == "" || startStr == "" || endStr == "" {
		c.JSON(http.StatusOK, common.Fail("roomNo, booker, start, end 不能为空"))
		return
	}

	// 解析时间
	layout := "2006-01-02 15:04:05"
	startAt, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail("start 时间格式错误，应为 \"2026-01-01 10:00:00\""))
		return
	}

	endAt, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail("end 时间格式错误，应为 \"2026-01-01 10:00:00\""))
		return
	}

	booking, err := h.bookingSvc.Book(c.Request.Context(), roomNo, booker, startAt, endAt)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(booking))
}

// Cancel POST /api/meeting/booking/cancel
// 取消预约
// 参数：bookingNo
func (h *BookingHandler) Cancel(c *gin.Context) {
	bookingNo := c.PostForm("bookingNo")
	if bookingNo == "" {
		c.JSON(http.StatusOK, common.Fail("bookingNo 不能为空"))
		return
	}

	if err := h.bookingSvc.Cancel(c.Request.Context(), bookingNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(nil))
}

// List GET /api/meeting/booking/list
// 按会议室和日期查询预约
// 参数：roomNo, date (格式: "2026-01-01"), pageNum, pageSize
func (h *BookingHandler) List(c *gin.Context) {
	roomNo := c.Query("roomNo")
	dateStr := c.Query("date")
	pageNumStr := c.Query("pageNum")
	pageSizeStr := c.Query("pageSize")

	if roomNo == "" || dateStr == "" {
		c.JSON(http.StatusOK, common.Fail("roomNo 和 date 不能为空"))
		return
	}

	// 解析日期
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail("date 格式错误，应为 \"2026-01-01\""))
		return
	}

	// 解析分页参数
	pageNum := 1
	if pageNumStr != "" {
		if n, err := strconv.Atoi(pageNumStr); err == nil && n > 0 {
			pageNum = n
		}
	}

	pageSize := 20
	if pageSizeStr != "" {
		if n, err := strconv.Atoi(pageSizeStr); err == nil && n > 0 {
			pageSize = n
		}
	}

	data, err := h.bookingSvc.ListByRoomAndDate(c.Request.Context(), roomNo, date, pageNum, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(data))
}