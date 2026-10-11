package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/gorm"
)

type ShareRecord struct {
}

// List 列表
// @Tags 分享记录
// @Summary 分享记录列表
// @Description 分享记录列表
// @Accept  json
// @Produce  json
// @Param user_id query int false "用户ID"
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/share_record/list [get]
// @Security token
func (sr *ShareRecord) List(c *gin.Context) {
	query := &admin.ShareRecordQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.ShareRecordService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.UserId > 0 {
			tx.Where("user_id = ?", query.UserId)
		}
	})
	response.Success(c, res)
}

// Delete 删除
// @Tags 分享记录
// @Summary 分享记录删除
// @Description 分享记录删除
// @Accept  json
// @Produce  json
// @Param body body admin.ShareRecordForm true "分享记录信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/share_record/delete [post]
// @Security token
func (sr *ShareRecord) Delete(c *gin.Context) {
	f := &admin.ShareRecordForm{}
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
	i := service.AllService.ShareRecordService.InfoById(f.Id)
	if i.Id > 0 {
		err := service.AllService.ShareRecordService.Delete(i)
		if err == nil {
			response.Success(c, nil)
			return
		}
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// BatchDelete 批量删除
// @Tags 分享记录
// @Summary 批量分享记录
// @Description 批量分享记录
// @Accept  json
// @Produce  json
// @Param body body admin.PeerShareRecordBatchDeleteForm true "id"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/share_record/batchDelete [post]
// @Security token
func (sr *ShareRecord) BatchDelete(c *gin.Context) {
	f := &admin.PeerShareRecordBatchDeleteForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if len(f.Ids) == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	// 批量删除：把被删分享对应的设备ID列表（去重）覆盖进 ids 键
	res := service.AllService.ShareRecordService.List(1, uint(len(f.Ids)), func(tx *gorm.DB) {
		tx.Where("id in (?)", f.Ids)
	})
	seen := map[string]bool{}
	peerIds := []string{}
	for _, i := range res.ShareRecords {
		if i.PeerId != "" && !seen[i.PeerId] {
			seen[i.PeerId] = true
			peerIds = append(peerIds, i.PeerId)
		}
	}
	if len(peerIds) > 0 {
		c.Set("op_log_extra", gin.H{"ids": peerIds})
	}
	err := service.AllService.ShareRecordService.BatchDelete(f.Ids)
	if err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, nil)
}
