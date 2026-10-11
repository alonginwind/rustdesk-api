package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/gorm"
)

type UserToken struct {
}

// List 列表
// @Tags 登录凭证
// @Summary 登录凭证列表
// @Description 登录凭证列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Param user_id query int false "用户ID"
// @Success 200 {object} response.Response{data=model.UserTokenList}
// @Failure 500 {object} response.Response
// @Router /admin/user_token/list [get]
// @Security token
func (ct *UserToken) List(c *gin.Context) {
	query := &admin.LoginTokenQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.UserService.TokenList(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.UserId > 0 {
			tx.Where("user_id = ?", query.UserId)
		}
		tx.Order("id desc")
	})
	response.Success(c, res)
}

// Delete 删除
// @Tags 登录凭证
// @Summary 登录凭证删除
// @Description 登录凭证删除
// @Accept  json
// @Produce  json
// @Param body body model.UserToken true "登录凭证信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user_token/delete [post]
// @Security token
func (ct *UserToken) Delete(c *gin.Context) {
	f := &model.UserToken{}
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
	l := service.AllService.UserService.TokenInfoById(f.Id)
	u := service.AllService.UserService.CurUser(c)
	if !service.AllService.UserService.IsAdmin(u) && l.UserId != u.Id {
		response.Fail(c, 101, response.TranslateMsg(c, "NoAccess"))
		return
	}
	if l.Id > 0 {
		// 补充操作日志详情：显示用户名
		extra := gin.H{}
		// 如果删除的是自己的token，直接用当前用户名；否则查询token所属用户
		if l.UserId == u.Id {
			extra["username"] = u.Username
		} else if tokenUser := service.AllService.UserService.InfoById(l.UserId); tokenUser.Id > 0 {
			extra["username"] = tokenUser.Username
		}
		c.Set("op_log_extra", extra)
		err := service.AllService.UserService.DeleteToken(l)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, err.Error())
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// BatchDelete 批量删除
// @Tags 登录凭证
// @Summary 登录凭证批量删除
// @Description 登录凭证批量删除
// @Accept  json
// @Produce  json
// @Param body body admin.UserTokenBatchDeleteForm true "登录凭证信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/user_token/batchDelete [post]
// @Security token
func (ct *UserToken) BatchDelete(c *gin.Context) {
	f := &admin.UserTokenBatchDeleteForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	ids := f.Ids
	if len(ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	// 补充操作日志详情：显示去重后的用户名列表
	res := service.AllService.UserService.TokenList(1, uint(len(ids)), func(tx *gorm.DB) {
		tx.Where("id in (?)", ids)
	})
	seen := map[uint]bool{}
	var userIds []uint
	for _, t := range res.UserTokens {
		if t.UserId > 0 && !seen[t.UserId] {
			seen[t.UserId] = true
			userIds = append(userIds, t.UserId)
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
	err := service.AllService.UserService.BatchDeleteUserToken(ids)
	if err == nil {
		response.Success(c, nil)
		return
	}
	response.Fail(c, 101, err.Error())
}
