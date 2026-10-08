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
	svc *service.MeetingService
}

func NewRoomHandler(svc *service.MeetingService) *RoomHandler {
	return &RoomHandler{svc: svc}
}

// CreateRoom 创建会议室
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	roomNo := c.Query("roomNo")
	if roomNo == "" {
		roomNo = c.PostForm("roomNo")
	}
	name := c.Query("name")
	if name == "" {
		name = c.PostForm("name")
	}
	capacityStr := c.Query("capacity")
	if capacityStr == "" {
		capacityStr = c.PostForm("capacity")
	}
	var capacity *int
	if capacityStr != "" {
		if n, err := strconv.Atoi(capacityStr); err == nil {
			capacity = &n
		}
	}

	room, err := h.svc.CreateRoom(c.Request.Context(), roomNo, name, capacity)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(room))
}

// DisableRoom 停用会议室
func (h *RoomHandler) DisableRoom(c *gin.Context) {
	roomNo := c.Query("roomNo")
	if roomNo == "" {
		roomNo = c.PostForm("roomNo")
	}

	if err := h.svc.DisableRoom(c.Request.Context(), roomNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(nil))
}

// BookingHandler 预约处理器
type BookingHandler struct {
	svc *service.MeetingService
}

func NewBookingHandler(svc *service.MeetingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// CreateBooking 创建预约
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	roomNo := c.Query("roomNo")
	if roomNo == "" {
		roomNo = c.PostForm("roomNo")
	}
	booker := c.Query("booker")
	if booker == "" {
		booker = c.PostForm("booker")
	}
	startStr := c.Query("start")
	if startStr == "" {
		startStr = c.PostForm("start")
	}
	endStr := c.Query("end")
	if endStr == "" {
		endStr = c.PostForm("end")
	}

	// 解析时间
	layout := "2006-01-02 15:04:05"
	startAt, err := time.Parse(layout, startStr)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail("开始时间格式错误，应为 " + layout))
		return
	}
	endAt, err := time.Parse(layout, endStr)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail("结束时间格式错误，应为 " + layout))
		return
	}

	booking, err := h.svc.CreateBooking(c.Request.Context(), &service.BookingArgs{
		RoomNo:  roomNo,
		Booker:  booker,
		StartAt: startAt,
		EndAt:   endAt,
	})
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(booking))
}

// CancelBooking 取消预约
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingNo := c.Query("bookingNo")
	if bookingNo == "" {
		bookingNo = c.PostForm("bookingNo")
	}

	if err := h.svc.CancelBooking(c.Request.Context(), bookingNo); err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(nil))
}

// ListBookings 查询预约列表
func (h *BookingHandler) ListBookings(c *gin.Context) {
	roomNo := c.Query("roomNo")
	dateStr := c.Query("date")
	pageNumStr := c.DefaultQuery("pageNum", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "20")

	pageNum, _ := strconv.Atoi(pageNumStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 解析日期
	layout := "2006-01-02"
	date, err := time.Parse(layout, dateStr)
	if err != nil {
		c.JSON(http.StatusOK, common.Fail("日期格式错误，应为 " + layout))
		return
	}

	result, err := h.svc.ListBookings(c.Request.Context(), &service.BookingListQuery{
		RoomNo:   roomNo,
		Date:     date,
		PageNum:  pageNum,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusOK, common.Fail(err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.OK(result))
}