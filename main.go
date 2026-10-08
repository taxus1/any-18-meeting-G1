package main

import (
	"log"

	"somepro/internal/config"
	"somepro/internal/router"
)

// somepro · Go 基座（Gin + GORM + Redis）。
// 启动：读环境变量（SERVER_PORT / DB_DSN / REDIS_ADDR）→ 起 HTTP 服务。
func main() {
	cfg := config.Load()

	db, err := config.OpenDB(cfg)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	rdb := config.OpenRedis(cfg)

	r := router.New(db, rdb)
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
