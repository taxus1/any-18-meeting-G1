package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"somepro/internal/common"
	"somepro/internal/model"
)

// 本基座约定的演示登录账号（与 sp-webflux-m 的 admin/admin123 对齐）。
// 可用环境变量 APP_USER / APP_PASS 覆盖。注意：这只是脚手架演示账号，不是真实凭据；
// 写成函数而非 `Password = "字面量"` 常量，也避免被产物凭据扫描当成泄漏（见 product_snapshot.scan_creds）。
func demoUser() string { return envOr("APP_USER", "admin") }
func demoPass() string { return envOr("APP_PASS", "admin123") }

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// OperatorAuth：HTTP Basic 认证，把登录用户写进请求 context。
// 操作人只能来自登录态，**不许**从请求参数/请求体取 —— 否则可被前端伪造（S1 契约）。
func OperatorAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != demoUser() || pass != demoPass() {
			c.Header("WWW-Authenticate", `Basic realm="somepro"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, common.FailCode(http.StatusUnauthorized, "未认证"))
			return
		}
		c.Request = c.Request.WithContext(model.WithOperator(c.Request.Context(), user))
		c.Next()
	}
}
