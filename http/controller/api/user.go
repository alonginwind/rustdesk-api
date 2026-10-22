package api

import (
	"github.com/gin-gonic/gin"
	apiResp "github.com/lejianwen/rustdesk-api/v2/http/response/api"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"net/http"
)

type User struct {
}

// currentUser 当前用户
// @Tags 用户
// @Summary 用户信息
// @Description 用户信息
// @Accept  json
// @Produce  json
// @Success 200 {object} apiResp.UserPayload
// @Failure 500 {object} response.Response
// @Router /currentUser [get]
// @Security token
//func (u *User) currentUser(c *gin.Context) {
//	user := service.AllService.UserService.CurUser(c)
//	up := (&apiResp.UserPayload{}).FromName(user)
//	c.JSON(http.StatusOK, up)
//}

// Info 用户信息
// @Tags 用户
// @Summary 用户信息
// @Description 用户信息，附带主控端策略（客户端每次启动经 token 鉴权调用，策略随响应下发）
// @Accept  json
// @Produce json
// @Success 200 {object} apiResp.CurrentUserRes
// @Failure 500 {object} response.Response
// @Router /currentUser [get]
// @Security token
func (u *User) Info(c *gin.Context) {
	user := service.AllService.UserService.CurUser(c)
	up := (&apiResp.UserPayload{}).FromUser(user)
	// 主控端策略随响应下发，覆盖 token 自动登录（免重新登录）的设备；
	// UserPayload 字段平铺保持与旧响应兼容，旧客户端忽略新增键
	clientStrategy, clientModifiedAt := ClientStrategyForPeer()
	c.JSON(http.StatusOK, apiResp.CurrentUserRes{
		UserPayload:      *up,
		ClientStrategy:   clientStrategy,
		ClientModifiedAt: clientModifiedAt,
	})
}
