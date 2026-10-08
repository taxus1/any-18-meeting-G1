package config

import (
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config 全部来自环境变量：容器/宿主两侧一致（宿主验收由 host_verify 注入同名变量）。
// 不要在代码里写死库名或地址。
type Config struct {
	Port          string
	DBDSN         string
	RedisAddr     string
	RedisPassword string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Load() Config {
	return Config{
		Port: env("SERVER_PORT", "8080"),
		// DSN 由运行环境给出（容器：host.docker.internal + 本侧独立库；宿主验收：127.0.0.1 + 本侧库）
		DBDSN:         env("DB_DSN", "root:root@tcp(127.0.0.1:3306)/some_pro?charset=utf8mb4&parseTime=true&loc=Local"),
		RedisAddr:     env("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
	}
}

func OpenDB(c Config) (*gorm.DB, error) {
	return gorm.Open(mysql.Open(c.DBDSN), &gorm.Config{})
}

func OpenRedis(c Config) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: c.RedisAddr, Password: c.RedisPassword})
}
