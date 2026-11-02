package admin_test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	admin "github.com/lejianwen/rustdesk-api/v2/http/controller/admin"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sessionPolicyFixture(t *testing.T) string {
	fixture(t)
	oldViper, oldConfig, oldExpiry := global.Viper, service.Config, global.Config.App.TokenExpire
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("app:\n  token-expire: 24h\n  register: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	global.Viper = viper.New()
	global.Viper.SetConfigFile(path)
	global.Config.App.TokenExpire = 24 * time.Hour
	service.Config = &global.Config
	t.Cleanup(func() { global.Viper = oldViper; service.Config = oldConfig; global.Config.App.TokenExpire = oldExpiry })
	return path
}

func policyAdmin(c *gin.Context) {
	u := service.AllService.UserService.CurUser(c)
	yes := true
	u.IsAdmin = &yes
	c.Set("curUser", u)
	(&admin.Config{}).SessionPolicy(c)
}

func TestSessionSettingsOnlyAdminCanReadOrSave(t *testing.T) {
	sessionPolicyFixture(t)
	for _, method := range []string{"GET", "POST"} {
		r := call(t, 2, (&admin.Config{}).SessionPolicy, method, "/policy", `{"duration":"720h","is_admin":true}`)
		if r["code"] != float64(403) {
			t.Fatalf("ordinary account accessed server configuration: %v", r)
		}
	}
	r := call(t, 1, policyAdmin, "GET", "/policy", "")
	success(t, r)
	data := r["data"].(map[string]interface{})
	if data["duration"] != "24h" {
		t.Fatal("did not read actual YAML value")
	}
}

func TestSessionSettingsSaveAndStaleRequest(t *testing.T) {
	path := sessionPolicyFixture(t)
	r := call(t, 1, policyAdmin, "GET", "/policy", "")
	success(t, r)
	revision := r["data"].(map[string]interface{})["revision"]
	payload, _ := json.Marshal(map[string]interface{}{"duration": "720h", "revision": revision})
	r = call(t, 1, policyAdmin, "POST", "/policy", string(payload))
	success(t, r)
	if r["data"].(map[string]interface{})["duration"] != "720h" {
		t.Fatal("saved setting not reflected")
	}
	r = call(t, 1, policyAdmin, "POST", "/policy", string(payload))
	if r["code"] == float64(0) {
		t.Fatal("stale settings form overwrote configuration")
	}
	if data, _ := os.ReadFile(path); len(data) == 0 {
		t.Fatal("configuration was lost")
	}
	for _, body := range []string{`{}`, `{"duration":"1h"}`, `{"duration":"0","revision":"stale"}`} {
		if r := call(t, 1, policyAdmin, "POST", "/policy", body); r["code"] == float64(0) {
			t.Fatal("invalid form accepted")
		}
	}
}
