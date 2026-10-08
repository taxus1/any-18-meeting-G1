package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// operatorKey 是 context 里携带「当前登录用户」的私有键。
// 操作人从登录态来（不在请求体里），由 middleware 写入请求 context，
// 仓储用 db.WithContext(ctx) 透传，GORM 钩子再从 tx.Statement.Context 取出 —— 这是本基座的 S1 契约。
type operatorKey struct{}

func WithOperator(ctx context.Context, op string) context.Context {
	return context.WithValue(ctx, operatorKey{}, op)
}

// OperatorFrom：取不到时回落 "system"（与鉴权中间件默认值一致）。
func OperatorFrom(ctx context.Context) string {
	if ctx == nil {
		return "system"
	}
	if v, ok := ctx.Value(operatorKey{}).(string); ok && v != "" {
		return v
	}
	return "system"
}

// Item 是 demo 限界上下文的聚合（示例，可删）。
// 审计字段由 GORM 钩子自动填充；软删除走 gorm.DeletedAt（查询自动追加 deleted_at IS NULL）。
type Item struct {
	ID         uint           `gorm:"primaryKey;column:id" json:"id"`
	Name       string         `gorm:"column:name;size:128;not null" json:"name"`
	Score      *int           `gorm:"column:score" json:"score"`
	CreateBy   string         `gorm:"column:create_by;size:64" json:"-"`
	CreateTime time.Time      `gorm:"column:create_time" json:"-"`
	UpdateBy   string         `gorm:"column:update_by;size:64" json:"-"`
	UpdateTime time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Item) TableName() string { return "t_demo_item" }

func (i *Item) BeforeCreate(tx *gorm.DB) error {
	op := OperatorFrom(tx.Statement.Context)
	now := time.Now()
	i.CreateBy, i.CreateTime = op, now
	i.UpdateBy, i.UpdateTime = op, now
	return nil
}

func (i *Item) BeforeUpdate(tx *gorm.DB) error {
	i.UpdateBy, i.UpdateTime = OperatorFrom(tx.Statement.Context), time.Now()
	return nil
}
