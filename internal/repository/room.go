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

//FindByRoomNo 通过 room_no 查找（含软删除）
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

//FindActiveByRoomNo 查找启用的会议室
func (r *RoomRepository) FindActiveByRoomNo(ctx context.Context, roomNo string) (*model.Room, error) {
	var room model.Room
	if err := r.db.WithContext(ctx).Where("room_no = ? AND deleted_at IS NULL AND status = 'ACTIVE'", roomNo).First(&room).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// DisableRoom 停用会议室
func (r *RoomRepository) DisableRoom(ctx context.Context, roomNo string) error {
	return r.db.WithContext(ctx).Model(&model.Room{}).
		Where("room_no = ? AND deleted_at IS NULL AND status = 'ACTIVE'", roomNo).
		Update("status", "DISABLED").Error
}

// BookingRepository 预约仓储
type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create 创建预约
func (r *BookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	return r.db.WithContext(ctx).Create(booking).Error
}

//FindByBookingNo 通过 booking_no 查找（含软删除）
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

// FindConflict 查找冲突的预约（同一会议室，时间段重叠）
// 重叠条件：start_at < :endAt AND end_at > :startAt
func (r *BookingRepository) FindConflict(ctx context.Context, roomID uint, startAt, endAt time.Time) (*model.Booking, error) {
	var booking model.Booking
	// 排除已取消的（deleted_at 不为空）
	query := r.db.WithContext(ctx).Where(
		"room_id = ? AND deleted_at IS NULL AND status = 'BOOKED' AND start_at < ? AND end_at > ?",
		roomID, endAt, startAt,
	).First(&booking)
	if query.Error != nil {
		if query.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, query.Error
	}
	return &booking, nil
}

// FindByRoomAndDate 查询某会议室某一天的预约（未取消）
// 返回分页结果和总条数
func (r *BookingRepository) FindByRoomAndDate(ctx context.Context, roomID uint, date time.Time, pageNum, pageSize int) ([]model.Booking, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	// 计算当天的起止时间
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// 总数
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("room_id = ? AND deleted_at IS NULL AND status = 'BOOKED' AND end_at > ? AND start_at < ?",
			roomID, endOfDay, startOfDay).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	var bookings []model.Booking
	if err := r.db.WithContext(ctx).Where(
		"room_id = ? AND deleted_at IS NULL AND status = 'BOOKED' AND end_at > ? AND start_at < ?",
		roomID, endOfDay, startOfDay).
		Order("start_at ASC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, total, nil
}

// Cancel 取消预约（软删除）
// 幂等：多次调用只取消一次
func (r *BookingRepository) Cancel(ctx context.Context, bookingNo string) error {
	// 只取消 status='BOOKED' 的，避免重复操作改变状态
	result := r.db.WithContext(ctx).Model(&model.Booking{}).
		Where("booking_no = ? AND deleted_at IS NULL AND status = 'BOOKED'", bookingNo).
		Update("status", "CANCELED")
	return result.Error
}

// FindByID 通过 ID 查找
func (r *BookingRepository) FindByID(ctx context.Context, id uint) (*model.Booking, error) {
	var booking model.Booking
	if err := r.db.WithContext(ctx).First(&booking, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}