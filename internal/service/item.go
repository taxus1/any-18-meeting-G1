package service

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"somepro/internal/model"
	"somepro/internal/repository"
)

type ItemService struct {
	repo *repository.ItemRepository
	rdb  *redis.Client
}

func NewItem(repo *repository.ItemRepository, rdb *redis.Client) *ItemService {
	return &ItemService{repo: repo, rdb: rdb}
}

// CreateItem：领域不变量（名称不能为空）在这里守，不要只靠 DB 的 not null。
func (s *ItemService) CreateItem(ctx context.Context, name string, score *int) (*model.Item, error) {
	if name == "" {
		return nil, errors.New("名称不能为空")
	}
	it := &model.Item{Name: name, Score: score}
	if err := s.repo.Create(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

func (s *ItemService) PageItems(ctx context.Context, pageNum, pageSize int, name string) (map[string]interface{}, error) {
	rows, total, err := s.repo.Page(ctx, pageNum, pageSize, name)
	if err != nil {
		return nil, err
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return map[string]interface{}{
		"content": rows, "total": total,
		"pageNum": pageNum, "pageSize": pageSize, "totalPages": totalPages,
	}, nil
}

// Cache：key 用字符串、value 用字符串，TTL 30s。跨模块读同一份要保证 key 前缀一致。
func (s *ItemService) Cache(ctx context.Context, key, value string) (string, error) {
	k := "demo:" + key
	if err := s.rdb.Set(ctx, k, value, 30*time.Second).Err(); err != nil {
		return "", err
	}
	return s.rdb.Get(ctx, k).Result()
}
