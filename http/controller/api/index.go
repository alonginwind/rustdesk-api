package api

import (
	"github.com/gin-gonic/gin"
	requstform "github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"net/http"
	"time"
)

type Index struct {
}

// Index 首页
// @Tags 首页
// @Summary 首页
// @Description 首页
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router / [get]
func (i *Index) Index(c *gin.Context) {
	response.Success(
		c,
		"Hello Gwen",
	)
}

// Heartbeat 心跳
// @Tags 首页
// @Summary 心跳
// @Description 心跳
// @Accept  json
// @Produce  json
// @Success 200 {object} nil
// @Failure 500 {object} response.Response
// @Router /heartbeat [post]
func (i *Index) Heartbeat(c *gin.Context) {
	info := &requstform.PeerInfoInHeartbeat{}
	err := c.ShouldBindJSON(info)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	if info.Uuid == "" {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	peer := service.AllService.PeerService.FindById(info.Id)
	if peer == nil || peer.RowId == 0 {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	peer.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(peer.Uuid, peer.Id)
	if peer.UserId == 0 || peer.Alias != "" {
		//如果在40s以内则不更新
		if time.Now().Unix()-peer.LastOnlineTime >= 30 {
			ab := service.AllService.AddressBookService.InfoByUserIdAndId(1, info.Id)//别名只同步全员地址簿，私人地址簿数据不同步
			var upp *model.Peer
			if ab == nil || ab.RowId == 0 {
				upp = &model.Peer{RowId: peer.RowId, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP()}
			} else {
				upp = &model.Peer{RowId: peer.RowId, Alias: ab.Alias, LastOnlineTime: time.Now().Unix(), LastOnlineIp: c.ClientIP()}
			}
			service.AllService.PeerService.Update(upp)
		}
	} else {//删除已登录的未绑定被控端
		service.AllService.PeerService.Delete(peer);
	}
	// 构建心跳响应，检查是否有策略配置需要下发给被控端
	// 对应 RustDesk 客户端 sync.rs 中心跳响应的 strategy 字段
	resp := gin.H{}
	deviceStrategy := service.AllService.PeerStrategyService.FindByPeerId(info.Id)
	defaultStrategy := service.AllService.PeerStrategyService.FindDefault()
	// 合并策略：设备策略 > 默认策略
	effectiveModifiedAt := service.AllService.PeerStrategyService.GetEffectiveModifiedAt(deviceStrategy, defaultStrategy)
	// 如果任一策略存在，且 modified_at 有变化，则下发合并后的配置
	hasStrategy := (deviceStrategy.Id > 0) || (defaultStrategy.Id > 0)
	strategyChanged := false
	var mergedConfig map[string]string
	if hasStrategy && info.ModifiedAt != effectiveModifiedAt {
		mergedConfig = service.AllService.PeerStrategyService.GetMergedConfigOptions(deviceStrategy, defaultStrategy)
		strategyChanged = true
	}

	// 检查地址簿预设信息是否有变化（集合名称、别名、主机名）
	abName, abAlias, abHostname := service.AllService.AddressBookService.GetPresetValuesForPeer(info.Id)
	presetChanged := abName != peer.PresetAbName || abAlias != peer.PresetAbAlias || abHostname != peer.PresetDevName
	if presetChanged {
		if !strategyChanged {
			if hasStrategy {
				mergedConfig = service.AllService.PeerStrategyService.GetMergedConfigOptions(deviceStrategy, defaultStrategy)
			} else {
				mergedConfig = map[string]string{}
			}
		}
		// 始终写入（包括空值），确保客户端能清除旧预设
		mergedConfig["preset-address-book-name"] = abName
		mergedConfig["preset-address-book-alias"] = abAlias
		mergedConfig["preset-device-name"] = abHostname
		// 更新 peer 表中的预设值，用于下次对比
		service.AllService.PeerService.UpdatePresets(peer.RowId, abName, abAlias, abHostname)
	}

	if strategyChanged || presetChanged {
		resp["strategy"] = gin.H{
			"config_options": mergedConfig,
		}
		resp["modified_at"] = effectiveModifiedAt
	}
	c.JSON(http.StatusOK, resp)
}

// Version 版本
// @Tags 首页
// @Summary 版本
// @Description 版本
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /version [get]
func (i *Index) Version(c *gin.Context) {
	//读取resources/version文件
	v := service.AllService.AppService.GetAppVersion()
	response.Success(
		c,
		v,
	)
}
