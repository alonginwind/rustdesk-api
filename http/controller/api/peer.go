package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	requstform "github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"net/http"
)

type Peer struct {
}

// SysInfo
// @Tags System
// @Summary 提交系统信息
// @Description 提交系统信息
// @Accept  json
// @Produce  json
// @Param body body requstform.PeerForm true "系统信息表单"
// @Success 200 {string} string "SYSINFO_UPDATED,ID_NOT_FOUND"
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo [post]
func (p *Peer) SysInfo(c *gin.Context) {
	f := &requstform.PeerForm{}
	err := c.ShouldBindBodyWith(f, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	fpe := f.ToPeer()
	pe := service.AllService.PeerService.FindById(f.Id)
	if pe.RowId == 0 {
		pe = f.ToPeer()
		pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		if pe.UserId == 0 || f.PresetAddressBookAlias != "" { //只同步未登录或者虽登录但手动注册的被控端
			pe.PresetAbAlias = f.PresetAddressBookAlias
			err = service.AllService.PeerService.Create(pe)
			if err != nil {
				response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
				return
			}
		} else {
			//已登录并且未手动注册的设备不入库，返回SYSINFO_UPDATED避免客户端每120秒重试
			c.String(http.StatusOK, "SYSINFO_UPDATED")
			return
		}
	} else {
		pe.UserId = service.AllService.UserService.FindLatestUserIdFromLoginLogByUuid(pe.Uuid, pe.Id)
		if pe.UserId == 0 || pe.Alias != "" { //只同步未登录或者虽登录但已绑定的被控端
			fpe.RowId = pe.RowId
			fpe.UserId = pe.UserId
			err = service.AllService.PeerService.Update(fpe)
			if err == nil {
				// Update 会忽略零值，这里显式同步 user_id（未登录时为0）
				err = service.AllService.PeerService.UpdateUserId(pe.RowId, pe.UserId)
			}
			if err != nil {
				response.Error(c, response.TranslateMsg(c, "OperationFailed")+err.Error())
				return
			}
		} else {
			//已登录并且未绑定的设备不入库，返回SYSINFO_UPDATED避免客户端每120秒重试
			c.String(http.StatusOK, "SYSINFO_UPDATED")
			return
		}
	}
	// 根据客户端上报的预设值自动将设备添加到地址簿
	service.AllService.AddressBookService.ApplyPresetToAddressBook(
		f.Id, f.Os, f.PresetAddressBookName, f.PresetAddressBookAlias, f.Hostname, f.Username,
	)

	//SYSINFO_UPDATED 上传成功
	//ID_NOT_FOUND 下次心跳会上传
	//直接响应文本
	c.String(http.StatusOK, "SYSINFO_UPDATED")
}

// SysInfoVer
// @Tags System
// @Summary 获取系统版本信息
// @Description 获取系统版本信息
// @Accept  json
// @Produce  json
// @Success 200 {string} string ""
// @Failure 500 {object} response.ErrorResponse
// @Router /sysinfo_ver [post]
func (p *Peer) SysInfoVer(c *gin.Context) {
	//读取resources/version文件
	v := service.AllService.AppService.GetAppVersion()
	// 加上启动时间，方便client上传信息
	v = fmt.Sprintf("%s\n%s", v, service.AllService.AppService.GetStartTime())
	c.String(http.StatusOK, v)
}
