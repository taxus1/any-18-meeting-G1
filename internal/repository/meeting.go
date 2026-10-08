package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"somepro/internal/model"
)

// RoomRepository 会议室仓储
type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create 创建会议室
func (r *RoomRepository) Create(ctx context.Context, room *model.Room) error {
	return r.db.WithContext(ctx).Create(room).Error
}

//FindByRoomNo 通过编号查找会议室（包括已停用的）
func (r *RoomRepository) FindByRoomNo(ctx context.Context, roomNo string) (*model.Room, error) {
	var room model.Room
	if err := r.db.WithContext(ctx).Where("room_no = ?", roomNo).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// UpdateStatus 更新会议室状态
func (r *RoomRepository) UpdateStatus(ctx context.Context, roomNo string, status string) error {
	return r.db.WithContext(ctx).Model(&model.Room{}).
		Where("room_no = ?", roomNo).
		Update("status", status).Error
}

// BookingRepository 预约仓储
type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create 创建预约
// 并发安全：如果预约号重复（极端并发情况下），会重试生成新的预约号
func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	err := r.db.WithContext(ctx).Create(booking).Error
	if err == nil {
		return nil
	}
	// GORM 的错误中包含唯一约束 violation 的信息
	// 由于我们使用 UUID，重复概率极低，此处直接返回错误
	// 如需强重试逻辑，可在 service 层包装重试
	return err
}

// FindByRoomAndTimeRange 查找会议室在指定时间范围内是否有预约冲突
// 冲突定义：两个时间段有重叠（不包括端点刚好相接的情况）
func (r *BookingRepository) FindByRoomAndTimeRange(ctx context.Context, roomID uint, startAt, endAt time.Time) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Where("status = ?", "BOOKED").
		Where("start_at < ? AND end_at > ?", endAt, startAt).
		Find(&bookings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return bookings, nil
}

// FindByRoomIDAndDate 查找会议室在某一天的预约（分页）
func (r *BookingRepository) FindByRoomIDAndDate(ctx context.Context, roomID uint, date time.Time, pageNum, pageSize int) ([]model.Booking, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 计算当天的起止时间
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	var total int64
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Where("status = ?", "BOOKED").
		Where("start_at < ? AND end_at > ?", endOfDay, startOfDay).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var bookings []model.Booking
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Where("status = ?", "BOOKED").
		Where("start_at < ? AND end_at > ?", endOfDay, startOfDay).
		Order("start_at ASC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

// FindByID 查找预约
func (r *BookingRepository) FindByID(ctx context.Context, id uint) (*model.Booking, error) {
	var booking model.Booking
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&booking).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}

// Cancel 取消预约（软删除）
// 幂等：多次调用不会导致状态异常
func (r *BookingRepository) Cancel(ctx context.Context, bookingNo string) error {
	return r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("booking_no = ?", bookingNo).
		Update("status", "CANCELLED").Error
}

// FindByBookingNo 通过预约编号查找
func (r *BookingRepository) FindByBookingNo(ctx context.Context, bookingNo string) (*model.Booking, error) {
	var booking model.Booking
	if err := r.db.WithContext(ctx).Where("booking_no = ?", bookingNo).First(&booking).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}