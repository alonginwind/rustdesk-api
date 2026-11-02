package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/http/response"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

func (co *Config) SessionPolicy(c *gin.Context) {
	us := service.AllService.UserService
	user := us.CurUser(c)
	if user == nil || user.IsAdmin == nil || !*user.IsAdmin {
		response.Fail(c, 403, "NoAccess")
		return
	}
	if c.Request.Method == "POST" {
		var form struct {
			Duration string `json:"duration"`
			Revision string `json:"revision"`
		}
		if err := c.ShouldBindJSON(&form); err != nil || form.Duration == "" || form.Revision == "" {
			response.Fail(c, 101, "ParamsError")
			return
		}
		if err := us.SaveSessionPolicy(form.Duration, form.Revision); err != nil {
			response.Fail(c, 101, err.Error())
			return
		}
	}
	policy, err := us.SessionPolicy()
	if err != nil {
		response.Fail(c, 101, "Cannot read session configuration")
		return
	}
	response.Success(c, policy)
}
