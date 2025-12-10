package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/http/response/api"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"time"
)

type WebClient struct {
}

// ServerConfig 服务配置
// @Tags WEBCLIENT
// @Summary 服务配置
// @Description 服务配置,给webclient提供api-server
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /server-config [get]
// @Security token
func (i *WebClient) ServerConfig(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)

	peers := map[string]*api.WebClientPeerPayload{}
	abs := service.AllService.AddressBookService.ListByUserIdAndCollectionId(u.Id, 0, 1, 100)
	for _, ab := range abs.AddressBooks {
		pp := &api.WebClientPeerPayload{}
		pp.FromAddressBook(ab)
		peers[ab.Id] = pp
	}
	response.Success(
		c,
		gin.H{
			"id_server": global.Config.Rustdesk.IdServer,
			"key":       global.Config.Rustdesk.Key,
			"peers":     peers,
		},
	)
}

// SharedPeer 分享的peer
// @Tags WEBCLIENT
// @Summary 分享的peer
// @Description 分享的peer
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /shared-peer [post]
func (i *WebClient) SharedPeer(c *gin.Context) {
	j := &gin.H{}
	c.ShouldBindJSON(j)
	t := (*j)["share_token"].(string)
	if t == "" {
		response.Fail(c, 101, "share_token is required")
		return
	}
	sr := service.AllService.AddressBookService.SharedPeer(t)
	if sr == nil || sr.Id == 0 {
		response.Fail(c, 101, "share not found")
		return
	}
	if sr.Expire != 0 {
		//判断是否过期,created_at + expire > now
		ca := time.Time(sr.CreatedAt)
		if ca.Add(time.Second * time.Duration(sr.Expire)).Before(time.Now()) {
			//过期删除记录
			service.AllService.DeleteShareByWebClientId(sr.PeerId)
			response.Fail(c, 101, "share expired")
			return
		}
	}

	ab := service.AllService.AddressBookService.InfoByUserIdAndId(sr.UserId, sr.PeerId)
	if ab.RowId == 0 {
		response.Fail(c, 101, "peer not found")
		return
	}
	pp := &api.WebClientPeerPayload{}
	pp.FromShareRecord(sr)
	pp.Info.Username = ab.Username
	pp.Info.Hostname = ab.Hostname
	response.Success(c, gin.H{
		"id_server": global.Config.Rustdesk.IdServer,
		"key":       global.Config.Rustdesk.Key,
		"peer":      pp,
	})
}

// QuerySharePeer 查询 ShareByWebClient PeerId
// @Tags WEBCLIENT
// @Summary 查询分享的 PeerId 是否存在
// @Description 查询 PeerId 是否在全局缓存中存在
// @Accept  json
// @Produce  json
// @Param peer_id query string true "PeerId"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /query-share-peer [get]
func (i *WebClient) QuerySharePeer(c *gin.Context) {
    peerId := c.Query("peer_id")
    if peerId == "" {
        response.Fail(c, 400, "peer_id is required")
        return
    }

    exists := service.AllService.AddressBookService.QueryShareByWebClientId(peerId)
    if !exists {
        response.Fail(c, 404, "peer not found")
        return
    }

	//登录后失效
	service.AllService.DeleteShareByWebClientId(peerId)
    response.Success(c, gin.H{"peer_id": peerId})
}

// ServerConfigV2 服务配置
// @Tags WEBCLIENT_V2
// @Summary 服务配置
// @Description 服务配置,给webclient提供api-server
// @Accept  json
// @Produce  json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /server-config-v2 [get]
// @Security token
func (i *WebClient) ServerConfigV2(c *gin.Context) {
	response.Success(
		c,
		gin.H{
			"id_server": global.Config.Rustdesk.IdServer,
			"key":       global.Config.Rustdesk.Key,
		},
	)
}
