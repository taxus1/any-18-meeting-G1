package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"somepro/internal/model"
	"somepro/internal/repository"
)

type RoomService struct {
	repo *repository.RoomRepository
}

func NewRoomService(repo *repository.RoomRepository) *RoomService {
	return &RoomService{repo: repo}
}

// CreateRoom 创建会议室
// roomNo: 唯一编号，不能重复
func (s *RoomService) CreateRoom(ctx context.Context, roomNo, name string, capacity int) (*model.Room, error) {
	// 检查会议室编号是否已存在
	existing, err := s.repo.FindByRoomNo(ctx, roomNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("会议室编号已存在")
	}

	room := &model.Room{
		RoomNo:   roomNo,
		Name:     name,
		Capacity: &capacity,
		Status:   "ACTIVE",
	}
	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

// DisableRoom 停用会议室
// 停用后不再接新预约，但已有预约保持不变
func (s *RoomService) DisableRoom(ctx context.Context, roomNo string) error {
	// 先检查是否存在
	room, err := s.repo.FindByRoomNo(ctx, roomNo)
	if err != nil {
		return err
	}
	if room == nil {
		return errors.New("会议室不存在")
	}

	// 更新状态为停用
	return s.repo.UpdateStatus(ctx, roomNo, "DISABLED")
}

type BookingService struct {
	roomRepo      *repository.RoomRepository
	bookingRepo   *repository.BookingRepository
	bookingNoMu   sync.Mutex // 预约号生成互斥锁
}

func NewBookingService(roomRepo *repository.RoomRepository, bookingRepo *repository.BookingRepository) *BookingService {
	return &BookingService{
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
	}
}

// generateBookingNo 生成唯一预约号
func (s *BookingService) generateBookingNo() string {
	// 使用 uuid 生成唯一编号
	return "BK" + uuid.New().String()[:22]
}

// Book 会议室预约
// 检查：会议室是否存在且启用、时间段不冲突、幂等性处理
func (s *BookingService) Book(ctx context.Context, roomNo, booker string, startAt, endAt time.Time) (*model.Booking, error) {
	// 参数校验
	if startAt.IsZero() || endAt.IsZero() {
		return nil, errors.New("开始或结束时间无效")
	}
	if !startAt.Before(endAt) {
		return nil, errors.New("开始时间必须早于结束时间")
	}
	if startAt.Before(time.Now().Add(-time.Hour)) {
		return nil, errors.New("不能预订过去的时间")
	}

	// 查找会议室
	room, err := s.roomRepo.FindByRoomNo(ctx, roomNo)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("会议室不存在")
	}
	if room.Status == "DISABLED" {
		return nil, errors.New("会议室已停用，无法预约")
	}

	// 检查时间段冲突
	conflicts, err := s.bookingRepo.FindByRoomAndTimeRange(ctx, room.ID, startAt, endAt)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, errors.New("该时间段已被预约，无法预订")
	}

	// 生成预约号并创建预约
	s.bookingNoMu.Lock()
	bookingNo := s.generateBookingNo()
	s.bookingNoMu.Unlock()

	booking := &model.Booking{
		BookingNo: bookingNo,
		RoomID:    room.ID,
		Booker:    booker,
		StartAt:   startAt,
		EndAt:     endAt,
		Status:    "BOOKED",
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

// Cancel 取消预约
// 幂等：多次取消不会报错，已取消的再次取消返回成功
func (s *BookingService) Cancel(ctx context.Context, bookingNo string) error {
	// 查找预约
	booking, err := s.bookingRepo.FindByBookingNo(ctx, bookingNo)
	if err != nil {
		return err
	}

	// 如果预约不存在，视为成功（幂等）
	if booking == nil {
		return nil
	}

	// 如果已经是取消状态，直接返回成功
	if booking.Status == "CANCELLED" {
		return nil
	}

	// 取消预约
	return s.bookingRepo.Cancel(ctx, bookingNo)
}

// ListByRoomAndDate 按会议室和日期查询预约
func (s *BookingService) ListByRoomAndDate(ctx context.Context, roomNo string, date time.Time, pageNum, pageSize int) (map[string]interface{}, error) {
	// 查找会议室
	room, err := s.roomRepo.FindByRoomNo(ctx, roomNo)
	if err != nil {
		return nil, err
	}
	if room == nil {
		// 会议室不存在，返回空列表
		return map[string]interface{}{
			"content": []model.Booking{},
			"total":   0,
			"pageNum": pageNum,
			"pageSize": pageSize,
		}, nil
	}

	// 查询预约
	bookings, total, err := s.bookingRepo.FindByRoomIDAndDate(ctx, room.ID, date, pageNum, pageSize)
	if err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	return map[string]interface{}{
		"content":    bookings,
		"total":      total,
		"pageNum":    pageNum,
		"pageSize":   pageSize,
		"totalPages": totalPages,
	}, nil
}

// ValidateBookingTime 检查时间段是否与已有预约冲突（供外部调用）
func (s *BookingService) ValidateBookingTime(ctx context.Context, roomID uint, startAt, endAt time.Time) error {
	conflicts, err := s.bookingRepo.FindByRoomAndTimeRange(ctx, roomID, startAt, endAt)
	if err != nil {
		return err
	}
	if len(conflicts) > 0 {
		return errors.New("该时间段已被预约")
	}
	return nil
}