package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/request/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/gorm"
)

type PeerStrategy struct{}

// List 被控端策略列表
// @Tags 被控端策略
// @Summary 被控端策略列表
// @Description 被控端策略列表
// @Accept  json
// @Produce  json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Param peer_id query string false "设备ID"
// @Success 200 {object} response.Response{data=model.PeerStrategyList}
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/list [get]
// @Security token
func (ct *PeerStrategy) List(c *gin.Context) {
	query := &admin.PeerStrategyQuery{}
	if err := c.ShouldBindQuery(query); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	res := service.AllService.PeerStrategyService.List(query.Page, query.PageSize, func(tx *gorm.DB) {
		tx.Where("peer_id != ''") // 排除默认策略
		if query.PeerId != "" {
			tx.Where("peer_id like ?", "%"+query.PeerId+"%")
		}
	})
	response.Success(c, res)
}

// Detail 被控端策略详情
// @Tags 被控端策略
// @Summary 被控端策略详情
// @Description 被控端策略详情
// @Accept  json
// @Produce  json
// @Param id path int true "ID"
// @Success 200 {object} response.Response{data=model.PeerStrategy}
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/detail/{id} [get]
// @Security token
func (ct *PeerStrategy) Detail(c *gin.Context) {
	id := c.Param("id")
	iid, _ := strconv.Atoi(id)
	ps := service.AllService.PeerStrategyService.Info(uint(iid))
	if ps.Id > 0 {
		response.Success(c, ps)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// Create 创建被控端策略
// @Tags 被控端策略
// @Summary 创建被控端策略
// @Description 创建被控端策略，设置被控端的配置参数
// @Accept  json
// @Produce  json
// @Param body body admin.PeerStrategyForm true "策略信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/create [post]
// @Security token
func (ct *PeerStrategy) Create(c *gin.Context) {
	f := &admin.PeerStrategyForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	// 检查是否已存在
	exist := service.AllService.PeerStrategyService.FindByPeerId(f.PeerId)
	if exist.Id > 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemExists"))
		return
	}
	ps := f.ToPeerStrategy()
	if err := service.AllService.PeerStrategyService.Create(ps); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, ps)
}

// Update 更新被控端策略
// @Tags 被控端策略
// @Summary 更新被控端策略
// @Description 更新被控端策略，更新后会在下次心跳时下发给被控端
// @Accept  json
// @Produce  json
// @Param body body admin.PeerStrategyForm true "策略信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/update [post]
// @Security token
func (ct *PeerStrategy) Update(c *gin.Context) {
	f := &admin.PeerStrategyForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	if f.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	errList := global.Validator.ValidStruct(c, f)
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	ps := service.AllService.PeerStrategyService.Info(f.Id)
	if ps.Id == 0 {
		response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
		return
	}
	ps.PeerId = f.PeerId
	ps.ConfigOptions = f.ToPeerStrategy().ConfigOptions
	if err := service.AllService.PeerStrategyService.Update(ps); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	response.Success(c, ps)
}

// Delete 删除被控端策略
// @Tags 被控端策略
// @Summary 删除被控端策略
// @Description 删除被控端策略
// @Accept  json
// @Produce  json
// @Param body body admin.PeerStrategyForm true "策略信息"
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/delete [post]
// @Security token
func (ct *PeerStrategy) Delete(c *gin.Context) {
	f := &admin.PeerStrategyForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	errList := global.Validator.ValidVar(c, f.Id, "required,gt=0")
	if len(errList) > 0 {
		response.Fail(c, 101, errList[0])
		return
	}
	ps := service.AllService.PeerStrategyService.Info(f.Id)
	if ps.Id > 0 {
		if err := service.AllService.PeerStrategyService.Delete(ps); err != nil {
			response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
			return
		}
		response.Success(c, nil)
		return
	}
	response.Fail(c, 101, response.TranslateMsg(c, "ItemNotFound"))
}

// Default 获取默认策略
// @Tags 被控端策略
// @Summary 获取默认策略
// @Description 获取全局默认策略，未单独配置策略的被控端会使用此默认策略
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response{data=model.PeerStrategy}
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/default [get]
// @Security token
func (ct *PeerStrategy) Default(c *gin.Context) {
	ps := service.AllService.PeerStrategyService.FindDefault()
	response.Success(c, ps)
}

// UpdateDefault 设置/更新默认策略
// @Tags 被控端策略
// @Summary 设置默认策略
// @Description 设置全局默认策略，未单独配置策略的被控端会在心跳时收到此配置
// @Accept  json
// @Produce  json
// @Param body body admin.DefaultStrategyForm true "策略信息"
// @Success 200 {object} response.Response{data=model.PeerStrategy}
// @Failure 500 {object} response.Response
// @Router /admin/peer_strategy/default/update [post]
// @Security token
func (ct *PeerStrategy) UpdateDefault(c *gin.Context) {
	f := &admin.DefaultStrategyForm{}
	if err := c.ShouldBindJSON(f); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	ps := f.ToPeerStrategy()
	if err := service.AllService.PeerStrategyService.CreateOrUpdateDefault(ps); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed")+err.Error())
		return
	}
	// 重新查询，确保返回完整记录（含 Id）
	ps = service.AllService.PeerStrategyService.FindDefault()
	response.Success(c, ps)
}
