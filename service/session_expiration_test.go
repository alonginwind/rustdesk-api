package service

import (
	"github.com/lejianwen/rustdesk-api/v2/config"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sessionConfigFixture(t *testing.T, text string) string {
	t.Helper()
	oldConfig, oldViper, oldRuntime := Config, global.Viper, runtimeSessionLifetime.Load()
	Config = &config.Config{App: config.App{TokenExpire: 24 * time.Hour}}
	runtimeSessionLifetime.Store(0)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	global.Viper = viper.New()
	global.Viper.SetConfigFile(path)
	t.Cleanup(func() { Config = oldConfig; global.Viper = oldViper; runtimeSessionLifetime.Store(oldRuntime) })
	return path
}

func TestSessionRenewalUsesDefaultAndNeverRevivesExpiredOrRevoked(t *testing.T) {
	sessionConfigFixture(t, "app: {}\n")
	Config.App.TokenExpire = 0
	oldDB := DB
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tokens.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	DB = db
	t.Cleanup(func() { sql, _ := db.DB(); sql.Close(); DB = oldDB })
	if err = db.AutoMigrate(&model.UserToken{}); err != nil {
		t.Fatal(err)
	}
	us := &UserService{}
	active := &model.UserToken{Token: "fixture-only-active", ExpiredAt: time.Now().Add(time.Minute).Unix()}
	if err = db.Create(active).Error; err != nil {
		t.Fatal(err)
	}
	us.AutoRefreshAccessToken(active)
	if active.ExpiredAt < time.Now().Add(167*time.Hour).Unix() {
		t.Fatal("default policy failed to renew")
	}
	expired := &model.UserToken{Token: "fixture-only-expired", ExpiredAt: time.Now().Add(-time.Minute).Unix()}
	db.Create(expired)
	before := expired.ExpiredAt
	us.AutoRefreshAccessToken(expired)
	if expired.ExpiredAt != before {
		t.Fatal("expired session revived")
	}
	if user, _ := us.InfoByAccessToken(expired.Token); user.Id != 0 {
		t.Fatal("expired session authenticated")
	}
	if err = us.DeleteToken(active); err != nil {
		t.Fatal(err)
	}
	if user, _ := us.InfoByAccessToken(active.Token); user.Id != 0 {
		t.Fatal("revoked session authenticated")
	}
}

func TestSessionExpiryDefaultIsSevenDays(t *testing.T) {
	sessionConfigFixture(t, "app: {}\n")
	Config.App.TokenExpire = 0
	now := time.Now().Unix()
	got := (&UserService{}).UserTokenExpireTimestamp() - now
	if got < 604799 || got > 604801 {
		t.Fatalf("default lifetime=%d seconds; want seven days", got)
	}
}

func TestSessionExpiryUsesConfiguredDuration(t *testing.T) {
	sessionConfigFixture(t, "app: {}\n")
	Config.App.TokenExpire = 720 * time.Hour
	now := time.Now().Unix()
	got := (&UserService{}).UserTokenExpireTimestamp() - now
	if got < 2591999 || got > 2592001 {
		t.Fatalf("configured lifetime=%d seconds", got)
	}
}

func TestSessionConfigRoundTripPreservesOtherFieldsAndComments(t *testing.T) {
	text := "# Keep header\napp:\n  token-expire: 24h # Keep session comment\n  register: false\njwt:\n  key: fixture-not-a-real-secret\ngin:\n  api-addr: 127.0.0.1:21114\n"
	path := sessionConfigFixture(t, text)
	us := &UserService{}
	current, err := us.SessionPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if current.Duration != "24h" || !current.Writable {
		t.Fatalf("wrong file state: %+v", current)
	}
	if err = us.SaveSessionPolicy("720h", current.Revision); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var beforeMap, afterMap map[string]interface{}
	yaml.Unmarshal([]byte(text), &beforeMap)
	yaml.Unmarshal(after, &afterMap)
	afterMap["app"].(map[string]interface{})["token-expire"] = "24h"
	beforeYaml, _ := yaml.Marshal(beforeMap)
	afterYaml, _ := yaml.Marshal(afterMap)
	if string(beforeYaml) != string(afterYaml) {
		t.Fatal("unrelated configuration changed")
	}
	for _, comment := range []string{"Keep header", "Keep session comment"} {
		if !strings.Contains(string(after), comment) {
			t.Fatal("comment lost")
		}
	}
	if us.UserTokenLifetime() != 720*time.Hour {
		t.Fatal("live policy did not update")
	}
	backups, _ := filepath.Glob(path + ".session-policy-*.bak")
	if len(backups) != 1 {
		t.Fatal("protected rollback copy missing")
	}
	saved, _ := os.ReadFile(backups[0])
	if string(saved) != text {
		t.Fatal("backup differs from original")
	}
	if err = us.SaveSessionPolicy("48h", current.Revision); err == nil {
		t.Fatal("stale browser revision accepted")
	}
}

func TestSessionConfigRejectsInvalidOrManagedChanges(t *testing.T) {
	path := sessionConfigFixture(t, "app:\n  token-expire: 24h\n")
	us := &UserService{}
	current, _ := us.SessionPolicy()
	for _, bad := range []string{"", "0", "-1h", "1ns", "999999h", "1h\nregister: true"} {
		if err := us.SaveSessionPolicy(bad, current.Revision); err == nil {
			t.Fatalf("invalid duration accepted: %q", bad)
		}
	}
	if data, _ := os.ReadFile(path); string(data) != "app:\n  token-expire: 24h\n" {
		t.Fatal("invalid edit changed file")
	}
	t.Setenv("RUSTDESK_API_APP_TOKEN_EXPIRE", "48h")
	current, _ = us.SessionPolicy()
	if !current.EnvironmentManaged || current.Writable {
		t.Fatal("environment control not reported")
	}
	if err := us.SaveSessionPolicy("720h", current.Revision); err == nil {
		t.Fatal("environment-managed config edited")
	}
}

func TestSessionConfigDetectsExternalEdits(t *testing.T) {
	path := sessionConfigFixture(t, "app:\n  token-expire: 24h\n")
	us := &UserService{}
	current, _ := us.SessionPolicy()
	changed := "app:\n  token-expire: 48h\n  register: false\n"
	os.WriteFile(path, []byte(changed), 0600)
	if err := us.SaveSessionPolicy("720h", current.Revision); err == nil {
		t.Fatal("external changes overwritten")
	}
	if data, _ := os.ReadFile(path); string(data) != changed {
		t.Fatal("external edit lost")
	}
}

func TestSessionExpiryPreservesExplicitNegativeDuration(t *testing.T) {
	sessionConfigFixture(t, "app: {}\n")
	Config.App.TokenExpire = -time.Hour
	if (&UserService{}).UserTokenExpireTimestamp() >= time.Now().Unix() {
		t.Fatal("explicitly expired policy was changed into a valid session")
	}
}

func TestSessionConfigDoesNotChangeSharedYamlAnchorValues(t *testing.T) {
	text := "app:\n  token-expire: &shared 24h\njwt:\n  expire-duration: *shared\n"
	path := sessionConfigFixture(t, text)
	us := &UserService{}
	current, err := us.SessionPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if err = us.SaveSessionPolicy("720h", current.Revision); err == nil {
		t.Fatal("editing the anchor would change the independent JWT setting")
	}
	if data, _ := os.ReadFile(path); string(data) != text {
		t.Fatal("shared configuration changed")
	}
}
