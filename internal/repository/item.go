package repository

import (
	"context"

	"gorm.io/gorm"

	"somepro/internal/model"
)

// ItemRepository：只认 model.Item（领域对象），不把 ORM 细节泄到 service/handler。
type ItemRepository struct {
	db *gorm.DB
}

func NewItem(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// 每个方法都 db.WithContext(ctx)：把请求 context 透传给 GORM 钩子（操作人审计靠它）。
func (r *ItemRepository) Create(ctx context.Context, it *model.Item) error {
	return r.db.WithContext(ctx).Create(it).Error
}

func (r *ItemRepository) FindByID(ctx context.Context, id uint) (*model.Item, error) {
	var it model.Item
	if err := r.db.WithContext(ctx).First(&it, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &it, nil
}

func (r *ItemRepository) Page(ctx context.Context, pageNum, pageSize int, name string) ([]model.Item, int64, error) {
	if pageNum < 1 {
		pageNum = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var total int64
	cq := r.db.WithContext(ctx).Model(&model.Item{})
	if name != "" {
		cq = cq.Where("name LIKE ?", "%"+name+"%")
	}
	if err := cq.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Item
	q := r.db.WithContext(ctx).Model(&model.Item{})
	if name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	if err := q.Order("id DESC").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// SoftDelete：GORM 在 deleted_at 上做软删除（UPDATE ... SET deleted_at=?），不是物理删除。
func (r *ItemRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Item{}, id).Error
}
