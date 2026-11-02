package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RustDesk decodes common-oidc/ as Future<List<dynamic>>. JSON null is a
// runtime type error, even when the surrounding HTTP response is successful.
func TestLoginOptionsClientCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name      string
		webSSO    bool
		providers []string
		want      []string
	}{
		{name: "password only", want: []string{}},
		{name: "web SSO", webSSO: true, want: []string{model.OauthTypeWebauth}},
		{name: "configured provider", providers: []string{model.OauthTypeGithub}, want: []string{model.OauthTypeGithub}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = sqlDB.Close() })
			if err := db.AutoMigrate(&model.Oauth{}); err != nil {
				t.Fatal(err)
			}
			for _, name := range tc.providers {
				if err := db.Create(&model.Oauth{Op: name}).Error; err != nil {
					t.Fatal(err)
				}
			}
			previousDB, previousServices, previousSSO := service.DB, service.AllService, global.Config.App.WebSso
			t.Cleanup(func() {
				service.DB, service.AllService, global.Config.App.WebSso = previousDB, previousServices, previousSSO
			})
			service.DB = db
			service.AllService = &service.Service{OauthService: &service.OauthService{}}
			global.Config.App.WebSso = tc.webSSO

			router := gin.New()
			router.GET("/api/login-options", (&Login{}).LoginOptions)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/login-options", nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", response.Code)
			}
			var options []string
			if err := json.Unmarshal(response.Body.Bytes(), &options); err != nil {
				t.Fatal(err)
			}
			if len(options) != len(tc.want)+1 || !strings.HasPrefix(options[0], "common-oidc/") {
				t.Fatalf("unexpected login options: %q", options)
			}
			var providers []map[string]string
			if err := json.Unmarshal([]byte(strings.TrimPrefix(options[0], "common-oidc/")), &providers); err != nil {
				t.Fatal(err)
			}
			if providers == nil {
				t.Fatal("common-oidc providers must be a JSON array, got null: native RustDesk throws FutureOr<List<dynamic>>")
			}
			got := make([]string, 0, len(providers))
			for _, provider := range providers {
				got = append(got, provider["name"])
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("providers = %q, want %q", got, tc.want)
			}
			for i, provider := range tc.want {
				if options[i+1] != "oidc/"+provider {
					t.Fatalf("legacy OIDC option changed: %q", options[i+1])
				}
			}
		})
	}
}
