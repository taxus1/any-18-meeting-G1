package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"somepro/internal/common"
	"somepro/internal/model"
)

// 本基座约定的登录账号（与 sp-webflux-m 的 admin/admin123 对齐）。
const (
	Username = "admin"
	Password = "admin123"
)

// OperatorAuth：HTTP Basic 认证，把登录用户写进请求 context。
// 操作人只能来自登录态，**不许**从请求参数/请求体取 —— 否则可被前端伪造（S1 契约）。
func OperatorAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != Username || pass != Password {
			c.Header("WWW-Authenticate", `Basic realm="somepro"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, common.FailCode(http.StatusUnauthorized, "未认证"))
			return
		}
		c.Request = c.Request.WithContext(model.WithOperator(c.Request.Context(), user))
		c.Next()
	}
}
