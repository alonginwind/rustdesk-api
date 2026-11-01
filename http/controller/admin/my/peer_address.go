package my

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

func (ct *Peer) ConnectionAddress(c *gin.Context) {
	if !global.Config.App.DeviceAddressBooks {
		response.Fail(c, 101, "DeviceAddressBooksDisabled")
		return
	}
	var f struct {
		RowId   uint    `json:"row_id"`
		Address *string `json:"connection_address"`
	}
	if err := c.ShouldBindJSON(&f); err != nil || f.RowId == 0 || f.Address == nil {
		response.Fail(c, 101, response.TranslateMsg(c, "ParamsError"))
		return
	}
	u := service.AllService.UserService.CurUser(c)
	if err := service.AllService.PeerService.SetConnectionAddress(f.RowId, *f.Address, &u.Id); err != nil {
		response.Fail(c, 101, response.TranslateMsg(c, "OperationFailed"))
		return
	}
	response.Success(c, nil)
}
