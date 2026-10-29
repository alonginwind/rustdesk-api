package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lejianwen/rustdesk-api/v2/global"
	admin "github.com/lejianwen/rustdesk-api/v2/http/controller/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/admin/my"
	api "github.com/lejianwen/rustdesk-api/v2/http/controller/api"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func fixture(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "books.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err = db.AutoMigrate(&model.User{}, &model.Peer{}, &model.AddressBook{}, &model.AddressBookCollection{}, &model.AddressBookCollectionRule{}, &model.PeerStrategy{}); err != nil {
		t.Fatal(err)
	}
	oldDB, oldServices, oldLogger := service.DB, service.AllService, global.Logger
	oldLocalizer, oldValidator := global.Localizer, global.Validator
	oldPath := global.Config.Gin.ResourcesPath
	service.DB = db
	service.AllService = &service.Service{UserService: &service.UserService{}, PeerService: &service.PeerService{}, AddressBookService: &service.AddressBookService{}, PeerStrategyService: &service.PeerStrategyService{}}
	global.Logger = logrus.New()
	global.Config.Gin.ResourcesPath = "../../../resources"
	global.InitI18n()
	global.ApiInitValidator()
	t.Cleanup(func() {
		service.DB, service.AllService, global.Logger = oldDB, oldServices, oldLogger
		global.Localizer, global.Validator = oldLocalizer, oldValidator
		global.Config.Gin.ResourcesPath = oldPath
	})
	for _, u := range []model.User{{IdModel: model.IdModel{Id: 1}, Username: "owner"}, {IdModel: model.IdModel{Id: 2}, Username: "other"}} {
		if err = db.Create(&u).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range []model.Peer{{RowId: 1, Id: "101", UserId: 1, Alias: "System name", Hostname: "device", AbId: 8}, {RowId: 2, Id: "102", UserId: 1}, {RowId: 3, Id: "103", UserId: 2}, {RowId: 4, Id: "104"}} {
		if err = db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []model.AddressBookCollection{{IdModel: model.IdModel{Id: 1}, UserId: 1, Name: "A"}, {IdModel: model.IdModel{Id: 2}, UserId: 1, Name: "B"}, {IdModel: model.IdModel{Id: 3}, UserId: 2, Name: "Private"}} {
		if err = db.Create(&c).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func call(t *testing.T, user uint, handler gin.HandlerFunc, method, path, body string) map[string]interface{} {
	t.Helper()
	r := gin.New()
	r.Handle(method, strings.Split(path, "?")[0], func(c *gin.Context) {
		c.Set("curUser", &model.User{IdModel: model.IdModel{Id: user}, GroupId: user})
		handler(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("HTTP %d: %s", w.Code, w.Body.String())
	}
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func success(t *testing.T, r map[string]interface{}) {
	t.Helper()
	if r["code"] != float64(0) {
		t.Fatalf("response: %v", r)
	}
}

func TestMyDevicesAreAccountDevices(t *testing.T) {
	fixture(t)
	for _, tc := range []struct {
		query string
		total int
	}{{"", 2}, {"&hostname=device", 1}, {"&id=102", 1}, {"&page_size=1", 2}} {
		r := call(t, 1, (&my.Peer{}).List, "GET", "/list?page=1"+tc.query, "")
		success(t, r)
		data := r["data"].(map[string]interface{})
		if int(data["total"].(float64)) != tc.total {
			t.Fatalf("query %s: %v", tc.query, data)
		}
		for _, raw := range data["list"].([]interface{}) {
			if raw.(map[string]interface{})["user_id"] != float64(1) {
				t.Fatal("foreign/unbound device exposed")
			}
		}
	}
}

func TestAdminAddPreservesOtherBooksAndPeer(t *testing.T) {
	db := fixture(t)
	for _, cid := range []int{1, 2, 2} {
		body := `{"id":"101","alias":"Book name","user_ids":[1],"collection_id":` + string(rune('0'+cid)) + `}`
		success(t, call(t, 1, (&admin.AddressBook{}).BatchCreate, "POST", "/add", body))
	}
	var count int64
	db.Model(&model.AddressBook{}).Where("id = ?", "101").Count(&count)
	if count != 2 {
		t.Fatalf("want one entry in both books after repeat add, got %d", count)
	}
	var p model.Peer
	db.First(&p, 1)
	if p.Alias != "System name" || p.AbId != 8 {
		t.Fatalf("book add changed peer: %+v", p)
	}
}

func TestAdminBookEditAndDeleteDoNotChangePeer(t *testing.T) {
	db := fixture(t)
	a := model.AddressBook{Id: "101", UserId: 1, CollectionId: 1, Alias: "A", Tags: []byte("[]")}
	db.Create(&a)
	b := model.AddressBook{Id: "101", UserId: 1, CollectionId: 2, Alias: "B", Tags: []byte("[]")}
	db.Create(&b)
	body, _ := json.Marshal(map[string]interface{}{"row_id": a.RowId, "id": "101", "user_id": 1, "collection_id": 1, "alias": "Edited", "hostname": "book host"})
	success(t, call(t, 1, (&admin.AddressBook{}).Update, "POST", "/edit", string(body)))
	del, _ := json.Marshal(map[string]interface{}{"row_id": a.RowId})
	success(t, call(t, 1, (&admin.AddressBook{}).Delete, "POST", "/delete", string(del)))
	var p model.Peer
	db.First(&p, 1)
	if p.Alias != "System name" || p.AbId != 8 || p.Hostname != "device" {
		t.Fatalf("book operation changed system/preset: %+v", p)
	}
	var surviving model.AddressBook
	if err := db.First(&surviving, b.RowId).Error; err != nil || surviving.Alias != "B" {
		t.Fatal("other book changed")
	}
}

func TestPersonalCreateCannotWriteAdminBook(t *testing.T) {
	db := fixture(t)
	r := call(t, 2, (&my.AddressBook{}).Create, "POST", "/add", `{"id":"unrelated","user_id":1,"collection_id":1,"alias":"Injected"}`)
	var count int64
	db.Model(&model.AddressBook{}).Where("user_id = ?", 1).Count(&count)
	if count != 0 {
		t.Fatalf("foreign book changed: %v", r)
	}
}

func TestLegacySyncDoesNotDeleteOtherCollections(t *testing.T) {
	db := fixture(t)
	a := model.AddressBook{Id: "101", UserId: 1, CollectionId: 1, Alias: "Keep A", Tags: []byte("[]")}
	db.Create(&a)
	b := model.AddressBook{Id: "101", UserId: 1, CollectionId: 2, Alias: "Keep B", Tags: []byte("[]")}
	db.Create(&b)
	if err := service.AllService.AddressBookService.UpdateAddressBook([]*model.AddressBook{}, 1); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&model.AddressBook{}).Count(&count)
	if count != 2 {
		t.Fatalf("legacy default-book sync removed other collections: %d", count)
	}
}

func TestPresetEnrollmentPreservesManualBooks(t *testing.T) {
	db := fixture(t)
	a := model.AddressBook{Id: "101", UserId: 1, CollectionId: 2, Alias: "Manual", Tags: []byte("[]")}
	db.Create(&a)
	service.AllService.AddressBookService.ApplyPresetToAddressBook(1, "101", "windows", "A", "Preset", "device", "user")
	var count int64
	db.Model(&model.AddressBook{}).Where("id = ?", "101").Count(&count)
	if count != 2 {
		t.Fatalf("preset removed manual membership: %d", count)
	}
}

func TestPeerAddressAndMemberships(t *testing.T) {
	db := fixture(t)
	if !db.Migrator().HasColumn(&model.Peer{}, "connection_address") || !db.Migrator().HasColumn(&model.AddressBook{}, "peer_row_id") {
		t.Fatal("additive address/link schema is missing")
	}
	db.Model(&model.Peer{}).Where("row_id = ?", 1).Update("connection_address", "100.64.1.2:21118")
	for _, b := range []map[string]interface{}{
		{"id": "book-local-target", "user_id": 1, "collection_id": 1, "alias": "Local A", "peer_row_id": 1, "tags": "[]"},
		{"id": "101", "user_id": 1, "collection_id": 2, "alias": "Local B", "peer_row_id": 0, "tags": "[]"},
		{"id": "private-target", "user_id": 2, "collection_id": 3, "alias": "Secret alias", "peer_row_id": 1, "tags": "[]"},
	} {
		if err := db.Table("address_books").Create(b).Error; err != nil {
			t.Fatal(err)
		}
	}
	r := call(t, 1, (&my.Peer{}).List, "GET", "/list?id=101", "")
	success(t, r)
	p := r["data"].(map[string]interface{})["list"].([]interface{})[0].(map[string]interface{})
	if p["connection_address"] != "100.64.1.2:21118" {
		t.Fatal("address override missing")
	}
	memberships, ok := p["book_memberships"].([]interface{})
	if !ok || len(memberships) != 2 {
		t.Fatalf("want both own aliases, no foreign alias: %v", p)
	}
	for _, cid := range []uint{4, 5} {
		if err := db.Create(&model.AddressBookCollection{IdModel: model.IdModel{Id: cid}, UserId: 1, Name: "New"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	success(t, call(t, 1, (&my.AddressBook{}).BatchCreateFromPeers, "POST", "/copy", `{"peer_ids":[1],"collection_ids":[4,5],"tags":[]}`))
	var entries []model.AddressBook
	db.Where("id = ?", "100.64.1.2:21118").Find(&entries)
	if len(entries) != 2 {
		t.Fatalf("copy must use default address in both books, got %d", len(entries))
	}
	// Entry destination override and a later device default must stay independent.
	body, _ := json.Marshal(map[string]interface{}{"row_id": entries[0].RowId, "id": "book-override", "user_id": 1, "collection_id": entries[0].CollectionId, "alias": "Override"})
	success(t, call(t, 1, (&my.AddressBook{}).Update, "POST", "/edit", string(body)))
	db.Model(&model.Peer{}).Where("row_id = ?", 1).Update("connection_address", "new-default")
	var edited map[string]interface{}
	db.Table("address_books").Where("row_id = ?", entries[0].RowId).Take(&edited)
	if edited["id"] != "book-override" || edited["peer_row_id"] != int64(1) {
		t.Fatalf("entry lost destination or device association: %v", edited)
	}
}

func TestBatchCopyRejectsForeignCollectionAtomically(t *testing.T) {
	db := fixture(t)
	r := call(t, 1, (&my.AddressBook{}).BatchCreateFromPeers, "POST", "/copy", `{"peer_ids":[1],"collection_ids":[1,3]}`)
	var count int64
	db.Model(&model.AddressBook{}).Count(&count)
	if r["code"] == float64(0) || count != 0 {
		t.Fatalf("foreign book accepted or partial write: %v, count=%d", r, count)
	}
}

func TestClientBookAliasStaysLocal(t *testing.T) {
	db := fixture(t)
	a := model.AddressBook{Id: "101", UserId: 1, CollectionId: 1, Alias: "A", Tags: []byte("[]")}
	db.Create(&a)
	r := gin.New()
	r.PUT("/book/:guid", func(c *gin.Context) {
		c.Set("curUser", &model.User{IdModel: model.IdModel{Id: 1}, GroupId: 1})
		(&api.Ab{}).PeerUpdate(c)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/book/1-1-1", strings.NewReader(`{"id":"101","alias":"Client alias","tags":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("client edit failed: %s", w.Body.String())
	}
	var p model.Peer
	db.First(&p, 1)
	if p.Alias != "System name" {
		t.Fatal("client alias changed system name")
	}
}

func TestConnectionAddressEditorPreservesIdentityAndExistingBooks(t *testing.T) {
	db := fixture(t)
	ab := model.AddressBook{Id: "old-book-target", UserId: 1, CollectionId: 1, Alias: "Local", Tags: []byte("[]")}
	db.Create(&ab)
	for _, address := range []string{"100.64.0.2:21118", ""} {
		body, _ := json.Marshal(map[string]interface{}{"row_id": 1, "connection_address": address})
		success(t, call(t, 1, (&my.Peer{}).ConnectionAddress, "POST", "/address", string(body)))
		var p model.Peer
		db.First(&p, 1)
		if p.Id != "101" || p.UserId != 1 || p.Alias != "System name" || p.AbId != 8 {
			t.Fatalf("identity or preset changed: %+v", p)
		}
		if address == "" && p.Address() != "101" {
			t.Fatal("empty override must restore ID default")
		}
	}
	var persisted model.AddressBook
	db.First(&persisted, ab.RowId)
	if persisted.Id != "old-book-target" {
		t.Fatal("default edit rewrote book")
	}
	for _, body := range []string{`{"row_id":3,"connection_address":"stolen"}`, `{"row_id":1}`, `{"row_id":1,"connection_address":"bad\naddress"}`} {
		if r := call(t, 1, (&my.Peer{}).ConnectionAddress, "POST", "/address", body); r["code"] == float64(0) {
			t.Fatalf("invalid/foreign address edit accepted: %s", body)
		}
	}
	success(t, call(t, 1, (&admin.Peer{}).ConnectionAddress, "POST", "/address", `{"row_id":4,"connection_address":"unbound-target"}`))
	if r := call(t, 1, (&admin.Peer{}).Update, "POST", "/update", `{"row_id":1,"id":"renamed"}`); r["code"] == float64(0) {
		t.Fatal("registered ID can be renamed from panel")
	}
}

func TestConcurrentCopiesKeepOneEntryAndLocalOverrides(t *testing.T) {
	db := fixture(t)
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- service.AllService.AddressBookService.CopyPeersToBooks([]uint{1}, []uint{1, 2}, 1, true, nil, nil, []byte("[]"))
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	db.Model(&model.AddressBook{}).Count(&count)
	if count != 2 {
		t.Fatalf("concurrent duplicate or lost book: %d", count)
	}
	var first model.AddressBook
	db.Where("collection_id = ?", 1).First(&first)
	body, _ := json.Marshal(map[string]interface{}{"row_id": first.RowId, "id": "local-target", "user_id": 1, "collection_id": 1, "alias": "Local"})
	success(t, call(t, 1, (&my.AddressBook{}).Update, "POST", "/edit", string(body)))
	if err := service.AllService.AddressBookService.CopyPeersToBooks([]uint{1}, []uint{1, 2}, 1, true, nil, nil, []byte("[]")); err != nil {
		t.Fatal(err)
	}
	db.Model(&model.AddressBook{}).Count(&count)
	if count != 2 {
		t.Fatal("retry duplicated a contact with an overridden destination")
	}
	var after model.AddressBook
	db.First(&after, first.RowId)
	if after.Id != "local-target" || after.Alias != "Local" {
		t.Fatal("retry overwrote local values")
	}
}

func TestClientAddToTwoBooksAndDeleteOnlyOne(t *testing.T) {
	db := fixture(t)
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("curUser", &model.User{IdModel: model.IdModel{Id: 1}, GroupId: 1}) })
	r.POST("/book/:guid", (&api.Ab{}).PeerAdd)
	r.DELETE("/book/:guid", (&api.Ab{}).PeerDel)
	native := func(method, path, body string) {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("native %s: %s", path, w.Body.String())
		}
	}
	native("POST", "/book/1-1-1", `{"id":"101","alias":"A","tags":[]}`)
	native("POST", "/book/1-1-2", `{"id":"101","alias":"B","tags":[]}`)
	native("POST", "/book/1-1-2", `{"id":"101","alias":"Must not overwrite B","tags":[]}`)
	native("DELETE", "/book/1-1-1", `["101"]`)
	var entries []model.AddressBook
	db.Where("user_id = ?", 1).Find(&entries)
	if len(entries) != 1 || entries[0].CollectionId != 2 || entries[0].Alias != "B" {
		t.Fatalf("native book isolation failed: %+v", entries)
	}
}

func TestAdditiveMigrationPreservesLegacyData(t *testing.T) {
	db := fixture(t)
	// Remove just the new fields to represent an existing v274 database. The
	// migration must add defaults without rewriting IDs, credentials or aliases.
	if err := db.Migrator().DropColumn(&model.Peer{}, "ConnectionAddress"); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrator().DropColumn(&model.AddressBook{}, "PeerRowId"); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO address_books (id,user_id,collection_id,alias,tags,password,hash) VALUES (?,?,?,?,?,?,?)", "legacy-target", 1, 1, "Keep alias", "[]", "fixture-password", "fixture-hash").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Peer{}, &model.AddressBook{}); err != nil {
		t.Fatal(err)
	}
	var old model.AddressBook
	db.Where("id = ?", "legacy-target").First(&old)
	if old.Id != "legacy-target" || old.Alias != "Keep alias" || old.Password != "fixture-password" || old.Hash != "fixture-hash" || old.PeerRowId != 0 {
		t.Fatal("migration rewrote existing contact")
	}
	var p model.Peer
	db.First(&p, 1)
	if p.Id != "101" || p.ConnectionAddress != "" || p.Address() != "101" || p.Alias != "System name" {
		t.Fatal("migration changed device identity/default")
	}
}

func TestPresetDoesNotOverwriteBookOverrides(t *testing.T) {
	db := fixture(t)
	db.Model(&model.Peer{}).Where("row_id = ?", 1).Update("connection_address", "device-default")
	service.AllService.AddressBookService.ApplyPresetToAddressBook(1, "101", "windows", "A", "Preset", "device", "user")
	var entry model.AddressBook
	db.Where("collection_id = ?", 1).First(&entry)
	if entry.Id != "device-default" || entry.PeerRowId != 1 {
		t.Fatalf("automatic enrollment ignored connection address: %+v", entry)
	}
	db.Model(&entry).Updates(map[string]interface{}{"id": "book-specific", "alias": "Manual"})
	service.AllService.AddressBookService.ApplyPresetToAddressBook(1, "101", "windows", "A", "Preset", "device", "user")
	var after model.AddressBook
	db.First(&after, entry.RowId)
	var count int64
	db.Model(&model.AddressBook{}).Where("collection_id = ?", 1).Count(&count)
	if count != 1 || after.Id != "book-specific" || after.Alias != "Manual" {
		t.Fatal("re-enrollment rewrote or duplicated local contact")
	}
}

func TestExplicitAssociationKeepsExistingDirectContact(t *testing.T) {
	db := fixture(t)
	ab := model.AddressBook{Id: "100.64.0.20:21118", UserId: 1, CollectionId: 1, Alias: "Direct contact", Tags: []byte("[]")}
	db.Create(&ab)
	body, _ := json.Marshal(map[string]interface{}{"row_id": ab.RowId, "id": ab.Id, "user_id": 1, "collection_id": 1, "alias": ab.Alias, "peer_row_id": 1})
	success(t, call(t, 1, (&admin.AddressBook{}).Update, "POST", "/edit", string(body)))
	var saved model.AddressBook
	db.First(&saved, ab.RowId)
	if saved.PeerRowId != 1 || saved.Id != ab.Id {
		t.Fatal("explicit device association lost or rewrote destination")
	}
	r := call(t, 1, (&my.Peer{}).List, "GET", "/list?id=101", "")
	success(t, r)
	p := r["data"].(map[string]interface{})["list"].([]interface{})[0].(map[string]interface{})
	if len(p["book_memberships"].([]interface{})) != 1 {
		t.Fatal("associated direct contact missing from aliases")
	}
}

func TestScopedMembershipsAndSharedReadRules(t *testing.T) {
	db := fixture(t)
	ab := model.AddressBook{Id: "private-target", PeerRowId: 1, UserId: 2, CollectionId: 3, Alias: "Private alias", Tags: []byte("[]"), Password: "fixture-secret", Hash: "fixture-hash"}
	db.Create(&ab)
	read := func(system bool) map[string]interface{} {
		t.Helper()
		h := (&my.Peer{}).List
		if system {
			h = (&admin.Peer{}).List
		}
		r := call(t, 1, h, "GET", "/list?id=101", "")
		success(t, r)
		return r["data"].(map[string]interface{})["list"].([]interface{})[0].(map[string]interface{})
	}
	if len(read(false)["book_memberships"].([]interface{})) != 0 {
		t.Fatal("private alias exposed")
	}
	if len(read(true)["book_memberships"].([]interface{})) != 1 {
		t.Fatal("system inventory missing admin-visible alias")
	}
	rule := model.AddressBookCollectionRule{UserId: 2, CollectionId: 3, Rule: model.ShareAddressBookRuleRuleRead, Type: model.ShareAddressBookRuleTypePersonal, ToId: 1}
	db.Create(&rule)
	visible := read(false)["book_memberships"].([]interface{})
	if len(visible) != 1 {
		t.Fatal("shared readable alias missing")
	}
	serialized, _ := json.Marshal(visible)
	if strings.Contains(string(serialized), "fixture-secret") || strings.Contains(string(serialized), "fixture-hash") {
		t.Fatal("membership response contains credentials")
	}
}

func TestBookDestinationCollisionDoesNotOverwriteEitherEntry(t *testing.T) {
	db := fixture(t)
	a := model.AddressBook{Id: "101", UserId: 1, CollectionId: 1, Alias: "A", Tags: []byte("[]")}
	db.Create(&a)
	b := model.AddressBook{Id: "target-b", UserId: 1, CollectionId: 1, Alias: "B", Tags: []byte("[]")}
	db.Create(&b)
	body, _ := json.Marshal(map[string]interface{}{"row_id": a.RowId, "id": b.Id, "user_id": 1, "collection_id": 1, "alias": "Collision"})
	if r := call(t, 1, (&my.AddressBook{}).Update, "POST", "/edit", string(body)); r["code"] == float64(0) {
		t.Fatal("duplicate destination accepted")
	}
	var after model.AddressBook
	db.First(&after, a.RowId)
	if after.Id != a.Id || after.Alias != "A" {
		t.Fatal("failed update changed source entry")
	}
}

func TestSystemAliasFilterMatchesAnyBookAlias(t *testing.T) {
	db := fixture(t)
	ab := model.AddressBook{Id: "different-target", PeerRowId: 1, UserId: 1, CollectionId: 1, Alias: "Unique local alias", Tags: []byte("[]")}
	db.Create(&ab)
	r := call(t, 1, (&admin.Peer{}).List, "GET", "/list?alias=Unique", "")
	success(t, r)
	if r["data"].(map[string]interface{})["total"] != float64(1) {
		t.Fatal("alias filter ignored book aliases")
	}
}

func TestCopyRollsBackOnStorageFailure(t *testing.T) {
	db := fixture(t)
	if err := db.Exec("CREATE TRIGGER fail_second_book BEFORE INSERT ON address_books WHEN NEW.collection_id = 2 BEGIN SELECT RAISE(ABORT, 'fixture failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	r := call(t, 1, (&my.AddressBook{}).BatchCreateFromPeers, "POST", "/copy", `{"peer_ids":[1],"collection_ids":[1,2]}`)
	var n int64
	db.Model(&model.AddressBook{}).Count(&n)
	if r["code"] == float64(0) || n != 0 {
		t.Fatalf("storage failure reported success or left a partial copy: %v / %d", r, n)
	}
}

func TestPersonalUpdateCannotRelinkToForeignDevice(t *testing.T) {
	db := fixture(t)
	a := model.AddressBook{Id: "external-target", UserId: 1, CollectionId: 1, Tags: []byte("[]")}
	db.Create(&a)
	body, _ := json.Marshal(map[string]interface{}{"row_id": a.RowId, "id": a.Id, "user_id": 1, "collection_id": 1, "peer_row_id": 3})
	if r := call(t, 1, (&my.AddressBook{}).Update, "POST", "/edit", string(body)); r["code"] == float64(0) {
		t.Fatal("personal route linked a foreign device")
	}
	body, _ = json.Marshal(map[string]interface{}{"row_id": a.RowId, "id": a.Id, "user_id": 2, "collection_id": 3})
	if r := call(t, 1, (&my.AddressBook{}).Update, "POST", "/edit", string(body)); r["code"] == float64(0) {
		t.Fatal("personal route transferred ownership")
	}
	success(t, call(t, 1, (&admin.AddressBook{}).Update, "POST", "/edit", string(body)))
	var after model.AddressBook
	db.First(&after, a.RowId)
	if after.UserId != 2 || after.CollectionId != 3 {
		t.Fatal("explicit admin transfer regressed")
	}
}
