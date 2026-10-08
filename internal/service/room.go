package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"somepro/internal/model"
	"somepro/internal/repository"
)

// MeetingService 会议室服务
type MeetingService struct {
	roomRepo    *repository.RoomRepository
	bookingRepo *repository.BookingRepository
	rdb         *redis.Client

	// 分布式锁 key 前缀（防止并发创建相同预约）
	bookingLockPrefix string
}

func NewMeetingService(roomRepo *repository.RoomRepository, bookingRepo *repository.BookingRepository, rdb *redis.Client) *MeetingService {
	return &MeetingService{
		roomRepo:          roomRepo,
		bookingRepo:       bookingRepo,
		rdb:               rdb,
		bookingLockPrefix: "meeting:booking:",
	}
}

// CreateRoom 创建会议室
func (s *MeetingService) CreateRoom(ctx context.Context, roomNo, name string, capacity *int) (*model.Room, error) {
	// 参数校验
	if roomNo == "" {
		return nil, errors.New("会议室编号不能为空")
	}
	if name == "" {
		return nil, errors.New("会议室名称不能为空")
	}

	// 检查编号是否已存在（包括已停用的）
	existing, err := s.roomRepo.FindByRoomNo(ctx, roomNo)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("会议室编号已存在")
	}

	room := &model.Room{
		RoomNo:   roomNo,
		Name:     name,
		Capacity: capacity,
		Status:   "ACTIVE",
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

// DisableRoom 停用会议室
func (s *MeetingService) DisableRoom(ctx context.Context, roomNo string) error {
	if roomNo == "" {
		return errors.New("会议室编号不能为空")
	}

	// 检查会议室是否存在且启用
	room, err := s.roomRepo.FindActiveByRoomNo(ctx, roomNo)
	if err != nil {
		return err
	}
	if room == nil {
		// 会议室不存在或已停用/删除，直接返回成功（幂等）
		return nil
	}

	if err := s.roomRepo.DisableRoom(ctx, roomNo); err != nil {
		return err
	}
	return nil
}

// BookingArgs 预约参数
type BookingArgs struct {
	RoomNo  string
	Booker  string
	StartAt time.Time
	EndAt   time.Time
}

// CreateBooking 创建预约
func (s *MeetingService) CreateBooking(ctx context.Context, args *BookingArgs) (*model.Booking, error) {
	// 参数校验
	if args.RoomNo == "" {
		return nil, errors.New("会议室编号不能为空")
	}
	if args.Booker == "" {
		return nil, errors.New("预订人不能为空")
	}
	if args.StartAt.IsZero() || args.EndAt.IsZero() {
		return nil, errors.New("开始时间和结束时间不能为空")
	}
	if args.StartAt.Compare(args.EndAt) >= 0 {
		return nil, errors.New("开始时间必须早于结束时间")
	}

	// 查找会议室（必须是启用状态）
	room, err := s.roomRepo.FindActiveByRoomNo(ctx, args.RoomNo)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("会议室不存在或已停用")
	}

	// 分布式锁，防止并发创建冲突的预约
	lockKey := s.bookingLockPrefix + args.RoomNo + ":" + args.StartAt.Format("20060102150405")
	// 尝试获取锁（最大等待 100ms）
	locked, err := s.tryAcquireLock(ctx, lockKey, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("获取锁失败: %w", err)
	}
	if !locked {
		return nil, errors.New("系统繁忙，请稍后重试")
	}
	defer s.releaseLock(ctx, lockKey)

	// 再次检查会议室状态（防止并发场景下会议室被停用）
	room, err = s.roomRepo.FindActiveByRoomNo(ctx, args.RoomNo)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, errors.New("会议室不存在或已停用")
	}

	// 检查是否有冲突的预约
	conflict, err := s.bookingRepo.FindConflict(ctx, room.ID, args.StartAt, args.EndAt)
	if err != nil {
		return nil, err
	}
	if conflict != nil {
		return nil, errors.New("该时间段已被预约")
	}

	// 生成 booking_no（格式：B20261008000001）
	bookingNo := s.generateBookingNo()

	booking := &model.Booking{
		BookingNo: bookingNo,
		RoomID:    room.ID,
		Booker:    args.Booker,
		StartAt:   args.StartAt,
		EndAt:     args.EndAt,
		Status:    "BOOKED",
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}
	return booking, nil
}

// tryAcquireLock 尝试获取分布式锁
func (s *MeetingService) tryAcquireLock(ctx context.Context, key string, timeout time.Duration) (bool, error) {
	// 使用 Redis SETNX 实现分布式锁
	// 这里简化实现：如果 Redis 不可用或未配置，返回 true（单实例场景）
	if s.rdb == nil {
		return true, nil
	}

	// 尝试在 timeout 内获取锁
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// SETNX + EXPIRE 原子操作：SET key value NX EX seconds
		err := s.rdb.SetNX(ctx, key, "1", 10*time.Second).Err()
		if err != nil {
			// Redis 错误，返回 false 表示获取锁失败
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// releaseLock 释放分布式锁
func (s *MeetingService) releaseLock(ctx context.Context, key string) {
	if s.rdb == nil {
		return
	}
	// 使用 Lua 脚本保证原子性
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`
	s.rdb.Eval(ctx, script, []string{key}, "1")
}

// generateBookingNo 生成预约编号
func (s *MeetingService) generateBookingNo() string {
	// 格式：B + 日期 + 序号
	// 简化实现：直接使用时间戳 + 随机数
	now := time.Now()
	return fmt.Sprintf("B%s%04d", now.Format("20060102150405"), now.Nanosecond()%10000)
}

// CancelBooking 取消预约
func (s *MeetingService) CancelBooking(ctx context.Context, bookingNo string) error {
	if bookingNo == "" {
		return errors.New("预约编号不能为空")
	}

	// 查找预约（包括已取消的）
	booking, err := s.bookingRepo.FindByBookingNo(ctx, bookingNo)
	if err != nil {
		return err
	}
	if booking == nil {
		// 预约不存在，返回成功（幂等）
		return nil
	}

	// 检查是否已取消
	if booking.Status == "CANCELED" {
		// 已取消，直接返回成功（幂等）
		return nil
	}

	// 取消预约
	if err := s.bookingRepo.Cancel(ctx, bookingNo); err != nil {
		return err
	}
	return nil
}

// BookingListQuery 预约列表查询参数
type BookingListQuery struct {
	RoomNo   string
	Date     time.Time
	PageNum  int
	PageSize int
}

// BookingInfo 预约信息
type BookingInfo struct {
	BookingNo string    `json:"bookingNo"`
	Booker    string    `json:"booker"`
	StartAt   time.Time `json:"startAt"`
	EndAt     time.Time `json:"endAt"`
}

// BookingListResponse 预约列表响应
type BookingListResponse struct {
	Content     []BookingInfo `json:"content"`
	Total       int64         `json:"total"`
	PageNum     int           `json:"pageNum"`
	PageSize    int           `json:"pageSize"`
	TotalPages  int           `json:"totalPages"`
}

// ListBookings 查询会议室某一天的预约
func (s *MeetingService) ListBookings(ctx context.Context, query *BookingListQuery) (*BookingListResponse, error) {
	// 参数校验
	if query.RoomNo == "" {
		return nil, errors.New("会议室编号不能为空")
	}
	if query.PageNum < 1 {
		query.PageNum = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 20
	}

	// 查找会议室
	room, err := s.roomRepo.FindByRoomNo(ctx, query.RoomNo)
	if err != nil {
		return nil, err
	}
	if room == nil {
		// 会议室不存在，返回空列表（不报错）
		return &BookingListResponse{
			Content:    []BookingInfo{},
			Total:      0,
			PageNum:    query.PageNum,
			PageSize:   query.PageSize,
			TotalPages: 0,
		}, nil
	}

	// 查询预约
	bookings, total, err := s.bookingRepo.FindByRoomAndDate(ctx, room.ID, query.Date, query.PageNum, query.PageSize)
	if err != nil {
		return nil, err
	}

	// 转换为返回格式
	infos := make([]BookingInfo, len(bookings))
	for i, b := range bookings {
		infos[i] = BookingInfo{
			BookingNo: b.BookingNo,
			Booker:    b.Booker,
			StartAt:   b.StartAt,
			EndAt:     b.EndAt,
		}
	}

	// 计算总页数
	totalPages := 0
	if query.PageSize > 0 && total > 0 {
		totalPages = int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	}

	return &BookingListResponse{
		Content:    infos,
		Total:      total,
		PageNum:    query.PageNum,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}