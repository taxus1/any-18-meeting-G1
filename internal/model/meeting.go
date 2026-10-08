package model

import (
	"time"

	"gorm.io/gorm"
)

// Room 会议室
// status: ACTIVE(启用) / DISABLED(停用)
type Room struct {
	ID         uint       `gorm:"primaryKey;column:id" json:"id"`
	RoomNo     string     `gorm:"column:room_no;size:32;not null;uniqueIndex" json:"roomNo"`
	Name       string     `gorm:"column:name;size:128;not null" json:"name"`
	Capacity   *int       `gorm:"column:capacity" json:"capacity"`
	Status     string     `gorm:"column:status;size:16;not null;default:ACTIVE" json:"status"`
	CreateBy   string     `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time  `gorm:"column:create_time" json:"-"`
	UpdateBy   string     `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time  `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Room) TableName() string { return "t_room" }

func (r *Room) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	r.CreateBy, r.CreateTime = op, now
	r.UpdateBy, r.UpdateTime = op, now
	return nil
}

func (r *Room) BeforeUpdate(tx *gorm.DB) error {
	r.UpdateBy, r.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}

// Booking 会议室预约
// status: BOOKED(已预约) / CANCELLED(已取消)
type Booking struct {
	ID         uint       `gorm:"primaryKey;column:id" json:"id"`
	BookingNo  string     `gorm:"column:booking_no;size:32;not null;uniqueIndex" json:"bookingNo"`
	RoomID     uint       `gorm:"column:room_id;not null" json:"roomId"`
	Booker     string     `gorm:"column:booker;size:64;not null" json:"booker"`
	StartAt    time.Time  `gorm:"column:start_at;not null" json:"startAt"`
	EndAt      time.Time  `gorm:"column:end_at;not null" json:"endAt"`
	Status     string     `gorm:"column:status;size:16;not null;default:BOOKED" json:"status"`
	CreateBy   string     `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time  `gorm:"column:create_time" json:"-"`
	UpdateBy   string     `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time  `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Booking) TableName() string { return "t_booking" }

func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	b.CreateBy, b.CreateTime = op, now
	b.UpdateBy, b.UpdateTime = op, now
	return nil
}

func (b *Booking) BeforeUpdate(tx *gorm.DB) error {
	b.UpdateBy, b.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}