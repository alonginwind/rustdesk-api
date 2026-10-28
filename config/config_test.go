package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	rustdeskapi "github.com/lejianwen/rustdesk-api/v2"
	"github.com/spf13/viper"
)

// An older config.yaml: no app.captcha-threshold, app.token-expire or jwt.key.
const oldConfig = `
app:
  register: true
admin:
  title: "From file"
jwt:
  private-key: ""
  expire-duration: 24h
postgresql:
  port: 5433
`

func initFromFile(t *testing.T, body string) *Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	viper.Reset()
	t.Cleanup(viper.Reset)
	c := &Config{}
	Init(c, path)
	return c
}

func shippedDefaults(t *testing.T) *viper.Viper {
	t.Helper()
	d := viper.New()
	d.SetConfigType("yaml")
	if err := d.ReadConfig(bytes.NewReader(rustdeskapi.DefaultConfigYAML)); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestInitMissingKeysUseShippedDefaults(t *testing.T) {
	d := shippedDefaults(t)
	c := initFromFile(t, oldConfig)

	if want := d.GetInt("app.captcha-threshold"); want == 0 || c.App.CaptchaThreshold != want {
		t.Errorf("app.captcha-threshold = %d, want shipped default %d", c.App.CaptchaThreshold, want)
	}
	if want := d.GetDuration("app.token-expire"); want == 0 || c.App.TokenExpire != want {
		t.Errorf("app.token-expire = %v, want shipped default %v", c.App.TokenExpire, want)
	}
	if want := d.GetString("jwt.key"); c.Jwt.Key != want {
		t.Errorf("jwt.key = %q, want shipped default %q", c.Jwt.Key, want)
	}
}

func TestInitFileOverridesDefaults(t *testing.T) {
	c := initFromFile(t, oldConfig)

	if !c.App.Register {
		t.Errorf("app.register = false, want true from file")
	}
	if c.Admin.Title != "From file" {
		t.Errorf("admin.title = %q, want %q from file", c.Admin.Title, "From file")
	}
	if c.Jwt.ExpireDuration != 24*time.Hour {
		t.Errorf("jwt.expire-duration = %v, want 24h from file", c.Jwt.ExpireDuration)
	}
	// The shipped default is the string "5432"; a YAML int in the file must still win.
	if c.Postgresql.Port != "5433" {
		t.Errorf("postgresql.port = %q, want %q from file", c.Postgresql.Port, "5433")
	}
}

func TestInitEnvOverridesFileAndDefaults(t *testing.T) {
	t.Setenv("RUSTDESK_API_ADMIN_TITLE", "From env")
	t.Setenv("RUSTDESK_API_APP_CAPTCHA_THRESHOLD", "-1")
	c := initFromFile(t, oldConfig)

	if c.Admin.Title != "From env" {
		t.Errorf("admin.title = %q, want %q from env over file", c.Admin.Title, "From env")
	}
	if c.App.CaptchaThreshold != -1 {
		t.Errorf("app.captcha-threshold = %d, want -1 from env over default", c.App.CaptchaThreshold)
	}
}

func TestInitMissingLdapAllowGroupStaysEmpty(t *testing.T) {
	c := initFromFile(t, `
ldap:
  enable: true
  user:
    admin-group: "cn=admins,dc=corp,dc=test"
`)

	if c.Ldap.User.AllowGroup != "" {
		t.Errorf("ldap.user.allow-group = %q, want empty when missing", c.Ldap.User.AllowGroup)
	}
	if c.Ldap.User.AdminGroup != "cn=admins,dc=corp,dc=test" {
		t.Errorf("ldap.user.admin-group = %q, want value from file", c.Ldap.User.AdminGroup)
	}
}

func TestInitEnvSetsMissingLdapAllowGroup(t *testing.T) {
	t.Setenv("RUSTDESK_API_LDAP_USER_ALLOW_GROUP", "cn=staff,dc=corp,dc=test")
	c := initFromFile(t, `
ldap:
  enable: true
`)

	if c.Ldap.User.AllowGroup != "cn=staff,dc=corp,dc=test" {
		t.Errorf("ldap.user.allow-group = %q, want value from env", c.Ldap.User.AllowGroup)
	}
}
