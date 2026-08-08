package middleware

import (
	"bytes"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

// 不需要记录日志的路径（登录类 + POST 查询类）
var skipOpLogPaths = map[string]bool{
	"/api/admin/login":         true,
	"/api/admin/logout":        true,
	"/api/admin/captcha":       true,
	"/api/admin/login-options": true,
	"/api/admin/oidc/auth":     true,
	// 以下 POST 接口实际是查询操作，不修改数据
	"/api/admin/peer/simpleData":  true,
	"/api/admin/user/myOauth":     true,
	"/api/admin/user/groupUsers":  true,
}

// OperationLog 操作日志中间件，记录所有 POST 写操作
// 必须在 BackendUserAuth 之后使用，以便获取当前用户信息
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录 POST 请求
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		// 跳过登录等接口
		if skipOpLogPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// 跳过操作日志自身的写入，防止递归
		if strings.Contains(c.Request.URL.Path, "/operation_log") {
			c.Next()
			return
		}

		// 读取请求体并还原
		var bodyStr string
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil {
			bodyStr = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// 截断过长的请求体（如上传图片）
		if len(bodyStr) > 2000 {
			bodyStr = bodyStr[:2000] + "...[truncated]"
		}

		// 执行后续处理
		c.Next()

		// 异步写入日志，不阻塞响应
		user := service.AllService.UserService.CurUser(c)
		if user == nil || user.Id == 0 {
			return
		}

		path := c.Request.URL.Path
		ip := c.ClientIP()
		resource, op := parsePath(path)

		go func() {
			_ = service.AllService.OperationLogService.Create(&model.OperationLog{
				UserId:   user.Id,
				Username: user.Username,
				Op:       op,
				Resource: resource,
				Detail:   bodyStr,
				Ip:       ip,
				Path:     path,
			})
		}()
	}
}

// parsePath 从 URL 路径中解析资源名和操作类型
// 例如 /api/admin/user/create → (user, create)
// 例如 /api/admin/peer/batchDelete → (peer, batch_delete)
func parsePath(path string) (resource, op string) {
	// 去掉 /api/admin/ 前缀
	p := strings.TrimPrefix(path, "/api/admin/")
	// 去掉 /my/ 前缀
	p = strings.TrimPrefix(p, "my/")

	parts := strings.Split(p, "/")
	if len(parts) >= 2 {
		resource = parts[0]
		op = parts[1]
	} else if len(parts) == 1 {
		resource = parts[0]
		op = "unknown"
	}

	// 统一操作名格式：camelCase → snake_case
	op = strings.ReplaceAll(op, "batchDelete", "batch_delete")
	op = strings.ReplaceAll(op, "batchCreate", "batch_create")
	op = strings.ReplaceAll(op, "batchUpdate", "batch_update")
	op = strings.ReplaceAll(op, "changeCurPwd", "change_cur_pwd")
	op = strings.ReplaceAll(op, "updatePassword", "update_password")

	return
}
