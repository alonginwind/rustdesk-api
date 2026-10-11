package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/gorm"
	"strconv"
)

type LoginLog struct {
}

// Detail 登录日志
// @Tags 登录日志
// @Summary 登录日志详情
// @Description 登录日志详情
// @Accept  json
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} response.Response{data=model.LoginLog}
// @Failure 500 {object} response.Response
// @Router /admin/login_log/detail/{id} [get]
// @Security token
func (ct *LoginLog) Detail(c *gin.Context) {
	id := c.Param("id")
	iid, _ := strconv.Atoi(id)
	u := service.AllService.LoginLogService.InfoById(uint(iid))
	if u.Id > 0 {
		response.Success(c, u)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
	return
}

// List 列表
// @Tags 登录日志
// @Summary 登录日志列表
// @Description 登录日志列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Param user_id query int false "用户ID"
// @Param device_id query string false "设备ID"
// @Success 200 {object} response.Response{data=model.LoginLogList}
// @Failure 500 {object} response.Response
// @Router /admin/login_log/list [get]
// @Security token
func (ct *LoginLog) List(c *gin.Context) {
	query := &admin.LoginLogQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.LoginLogService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.UserId > 0 {
			tx.Where("user_id = ?", query.UserId)
		}
		if query.DeviceId != "" {
			tx.Where("device_id like ?", "%"+query.DeviceId+"%")
		}
		tx.Order("id desc")
	})
	response.Success(c, res)
}

// Delete 删除
// @Tags 登录日志
// @Summary 登录日志删除
// @Description 登录日志删除
// @Accept  json
// @Produce  json
// @Param body body model.LoginLog true "登录日志信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/login_log/delete [post]
// @Security token
func (ct *LoginLog) Delete(c *gin.Context) {
	f := &model.LoginLog{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	id := f.Id
	errList := global.Validator.ValidVar(c, id, "required,gt=0")
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	l := service.AllService.LoginLogService.InfoById(f.Id)
	if l.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}
	// 删除请求体只有 id（日志行ID）：设备端登录时 id 覆盖为设备ID；webadmin 登录无设备则显示 "webadmin" 字样
	extra := gin.H{}
	if l.DeviceId != "" {
		extra["id"] = l.DeviceId
	} else {
		extra["id"] = "webadmin"
	}
	if u := service.AllService.UserService.InfoById(l.UserId); u.Id > 0 {
		extra["username"] = u.Username
	}
	c.Set("op_log_extra", extra)
	err := service.AllService.LoginLogService.Delete(l)
	if err == nil {
		response.Success(c, nil)
		return
	}
	response.Fail(c, 101, err.Error())
}

// BatchDelete 删除
// @Tags 登录日志
// @Summary 登录日志批量删除
// @Description 登录日志批量删除
// @Accept  json
// @Produce  json
// @Param body body admin.LoginLogIds true "登录日志"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/login_log/batchDelete [post]
// @Security token
func (ct *LoginLog) BatchDelete(c *gin.Context) {
	f := &admin.LoginLogIds{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if len(f.Ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	// 批量删除：把去重后的用户名覆盖进 ids 键，描述只显示用户名、不显示行ID列表
	res := service.AllService.LoginLogService.List(1, uint(len(f.Ids)), func(tx *gorm.DB) {
		tx.Where("id in (?)", f.Ids)
	})
	seen := map[uint]bool{}
	var userIds []uint
	for _, l := range res.LoginLogs {
		if l.UserId > 0 && !seen[l.UserId] {
			seen[l.UserId] = true
			userIds = append(userIds, l.UserId)
		}
	}
	names := []string{}
	if len(userIds) > 0 {
		for _, u := range service.AllService.UserService.ListByIds(userIds) {
			names = append(names, u.Username)
		}
	}
	if len(names) > 0 {
		c.Set("op_log_extra", gin.H{"ids": names})
	}
	err := service.AllService.LoginLogService.BatchDelete(f.Ids)
	if err == nil {
		response.Success(c, nil)
		return
	}
	response.Fail(c, 101, err.Error())
	return
}
