package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/lejianwen/rustdesk-api/v2/global"
	request "github.com/lejianwen/rustdesk-api/v2/http/request/api"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"net/http"
	"time"
)

type Audit struct {
}

// auditStored answers a post that got stored. A 2xx body that is not empty reads
// to the client as an unknown outcome, and it retries -- every retry that the
// nonce check below does not catch lands as one more record. The server the
// clients were written against answers with an empty body, so do the same.
func auditStored(c *gin.Context) {
	c.Status(http.StatusOK)
}

// auditFailed keeps a rejected post retryable: the client retries a 2xx whose
// body is {"error": ...}. Nothing was stored and no nonce was recorded, so the
// next attempt stores the record instead of it being lost.
func auditFailed(c *gin.Context, err error) {
	global.Logger.Warn("audit", err)
	c.JSON(http.StatusOK, response.ErrorResponse{Error: err.Error()})
}

// AuditConn
// @Tags 审计
// @Summary 审计连接
// @Description 审计连接
// @Accept  json
// @Produce  json
// @Param body body request.AuditConnForm true "审计连接"
// @Success 200 {string} string ""
// @Failure 500 {object} response.Response
// @Router /audit/conn [post]
func (a *Audit) AuditConn(c *gin.Context) {
	af := &request.AuditConnForm{}
	err := c.ShouldBindBodyWith(af, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	/*ttt := &gin.H{}
	c.ShouldBindBodyWith(ttt, binding.JSON)
	fmt.Println(ttt)*/
	ac := af.ToAuditConn()
	if af.Action == model.AuditActionNew {
		err = service.AllService.AuditService.CreateAuditConnIfNonceUnique(ac)
	} else if af.Action == model.AuditActionClose {
		ex := service.AllService.AuditService.InfoByPeerIdAndConnId(af.Id, af.ConnId)
		if ex.Id != 0 {
			ex.CloseTime = time.Now().Unix()
			err = service.AllService.AuditService.UpdateAuditConn(ex)
		}
	} else if af.Action == "" {
		ex := service.AllService.AuditService.InfoByPeerIdAndConnId(af.Id, af.ConnId)
		if ex.Id != 0 {
			up := &model.AuditConn{
				IdModel:     model.IdModel{Id: ex.Id},
				FromPeer:    ac.FromPeer,
				FromName:    ac.FromName,
				SessionId:   ac.SessionId,
				Type:        ac.Type,
				PrimaryAuth: ac.PrimaryAuth,
				TwoFactor:   ac.TwoFactor,
			}
			err = service.AllService.AuditService.UpdateAuditConn(up)
		}
	}
	if err != nil {
		auditFailed(c, err)
		return
	}
	auditStored(c)
}

// AuditFile
// @Tags 审计
// @Summary 审计文件
// @Description 审计文件
// @Accept  json
// @Produce  json
// @Param body body request.AuditFileForm true "审计文件"
// @Success 200 {string} string ""
// @Failure 500 {object} response.Response
// @Router /audit/file [post]
func (a *Audit) AuditFile(c *gin.Context) {
	aff := &request.AuditFileForm{}
	err := c.ShouldBindBodyWith(aff, binding.JSON)
	if err != nil {
		response.Error(c, response.TranslateMsg(c, "ParamsError")+err.Error())
		return
	}
	//ttt := &gin.H{}
	//c.ShouldBindBodyWith(ttt, binding.JSON)
	//fmt.Println(ttt)
	af := aff.ToAuditFile()
	err = service.AllService.AuditService.CreateAuditFileIfNonceUnique(af)
	if err != nil {
		auditFailed(c, err)
		return
	}
	auditStored(c)
}
