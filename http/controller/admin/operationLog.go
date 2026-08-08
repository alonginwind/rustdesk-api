package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/gorm"
)

type OperationLog struct {
}

// List 列表
// @Tags 操作日志
// @Summary 操作日志列表
// @Description 操作日志列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Param username query string false "操作人"
// @Param resource query string false "资源类型"
// @Param op query string false "操作类型"
// @Success 200 {object} response.Response{data=model.OperationLogList}
// @Failure 500 {object} response.Response
// @Router /admin/operation_log/list [get]
// @Security token
func (ct *OperationLog) List(c *gin.Context) {
	query := &admin.OperationLogQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.OperationLogService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		if query.Username != "" {
			tx.Where("username like ?", "%"+query.Username+"%")
		}
		if query.Resource != "" {
			tx.Where("resource = ?", query.Resource)
		}
		if query.Op != "" {
			tx.Where("op = ?", query.Op)
		}
	})
	response.Success(c, res)
}
