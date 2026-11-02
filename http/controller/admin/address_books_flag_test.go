package admin_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/lejianwen/rustdesk-api/v2/global"
	admin "github.com/lejianwen/rustdesk-api/v2/http/controller/admin"
	"github.com/lejianwen/rustdesk-api/v2/http/controller/admin/my"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/service"
)

func legacyMode(t *testing.T) {
	t.Helper()
	flag := reflect.ValueOf(&global.Config.App).Elem().FieldByName("DeviceAddressBooks")
	if flag.IsValid() {
		flag.SetBool(false)
	}
}

func TestDisabledSchemeRestoresLegacyInventory(t *testing.T) {
	fixture(t)
	legacyMode(t)
	r := call(t, 1, (&my.Peer{}).List, "GET", "/list?page=1", "")
	success(t, r)
	data := r["data"].(map[string]interface{})
	if data["total"] != float64(1) || data["list"].([]interface{})[0].(map[string]interface{})["id"] != "104" {
		t.Fatalf("expected original unbound/no-alias selection: %v", data)
	}
}

func TestDisabledSchemeRestoresLegacyCleanupAndAlias(t *testing.T) {
	db := fixture(t)
	legacyMode(t)
	for _, body := range []string{
		`{"id":"101","alias":"First","user_ids":[1],"collection_id":1}`,
		`{"id":"101","alias":"Second","user_ids":[1],"collection_id":2}`,
	} {
		success(t, call(t, 1, (&admin.AddressBook{}).BatchCreate, "POST", "/add", body))
	}
	var n int64
	db.Model(&model.AddressBook{}).Where("id = ?", "101").Count(&n)
	p := service.AllService.PeerService.InfoByRowId(1)
	if n != 1 || p.Alias != "Second" {
		t.Fatalf("legacy cleanup/alias missing: count=%d alias=%q", n, p.Alias)
	}
}

func TestDisabledSchemeRejectsAddressEditor(t *testing.T) {
	db := fixture(t)
	legacyMode(t)
	for _, handler := range []struct{ personal bool }{{false}, {true}} {
		var r map[string]interface{}
		if handler.personal {
			r = call(t, 1, (&my.Peer{}).ConnectionAddress, "POST", "/edit", `{"row_id":1,"connection_address":"changed"}`)
		} else {
			r = call(t, 1, (&admin.Peer{}).ConnectionAddress, "POST", "/edit", `{"row_id":1,"connection_address":"changed"}`)
		}
		if r["code"] == float64(0) {
			t.Fatal("disabled editor accepted a write")
		}
	}
	var p model.Peer
	db.First(&p, 1)
	if p.ConnectionAddress != "" {
		t.Fatal("disabled editor changed stored address")
	}
}

func TestDisabledSchemeRestoresLegacySnapshotScope(t *testing.T) {
	db := fixture(t)
	legacyMode(t)
	ab := model.AddressBook{Id: "101", UserId: 1, CollectionId: 1, Alias: "Named", Tags: []byte("[]")}
	db.Create(&ab)
	if err := service.AllService.AddressBookService.UpdateAddressBook(nil, 1); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&model.AddressBook{}).Where("user_id = 1").Count(&n)
	if n != 0 {
		t.Fatalf("legacy full snapshot retained %d contacts", n)
	}
}

func TestPersonalOwnershipDoesNotDependOnScheme(t *testing.T) {
	fixture(t)
	legacyMode(t)
	r := call(t, 2, (&my.AddressBook{}).Create, "POST", "/create", `{"id":"101","user_id":1,"collection_id":1}`)
	if r["code"] == float64(0) {
		t.Fatal("legacy flag must not authorize writing another user's book")
	}
}

func TestDeploymentFlagIsReadOnlyAppCapability(t *testing.T) {
	fixture(t)
	for _, enabled := range []bool{false, true} {
		reflect.ValueOf(&global.Config.App).Elem().FieldByName("DeviceAddressBooks").SetBool(enabled)
		r := call(t, 1, (&admin.Config{}).AppConfig, "GET", "/app", "")
		success(t, r)
		if r["data"].(map[string]interface{})["device_address_books"] != enabled {
			t.Fatal("capability differs from server configuration")
		}
	}
}

func TestSwitchingOffAndOnRetainsNewSchemeData(t *testing.T) {
	db := fixture(t)
	success(t, call(t, 1, (&admin.Peer{}).ConnectionAddress, "POST", "/address", `{"row_id":1,"connection_address":"default-target"}`))
	success(t, call(t, 1, (&my.AddressBook{}).BatchCreateFromPeers, "POST", "/copy", `{"peer_ids":[1],"collection_ids":[1,2]}`))
	before := service.AllService.AddressBookService.InfoByUserIdAndIdAndCid(1, "default-target", 1)
	legacyMode(t)
	ab := service.AllService.AddressBookService.FromPeer(service.AllService.PeerService.InfoByRowId(1))
	if ab.Id != "101" || ab.PeerRowId != 0 {
		t.Fatalf("legacy copy uses new address: %+v", ab)
	}
	r := call(t, 1, (&admin.Peer{}).List, "GET", "/list?page_size=100", "")
	success(t, r)
	reflect.ValueOf(&global.Config.App).Elem().FieldByName("DeviceAddressBooks").SetBool(true)
	var after model.AddressBook
	db.First(&after, before.RowId)
	if after.PeerRowId != 1 || after.Id != "default-target" {
		t.Fatal("toggling removed stored association/address")
	}
	p := service.AllService.PeerService.InfoByRowId(1)
	if p.Address() != "default-target" {
		t.Fatal("toggling erased device default")
	}
	if err := service.AllService.PeerService.DecorateBooks([]*model.Peer{p}, nil, true); err != nil {
		t.Fatal(err)
	}
	if len(p.BookMemberships) != 2 {
		t.Fatalf("lost memberships: %+v", p.BookMemberships)
	}
}

func TestDisabledSchemeRestoresLegacyPresetEnrollment(t *testing.T) {
	db := fixture(t)
	legacyMode(t)
	db.Create(&model.AddressBook{Id: "101", UserId: 1, CollectionId: 2, Alias: "Manual", Tags: []byte("[]")})
	service.AllService.AddressBookService.ApplyPresetToAddressBook(1, "101", "windows", "A", "Preset", "preset-host", "preset-user")
	var n int64
	db.Model(&model.AddressBook{}).Where("id = ?", "101").Count(&n)
	p := service.AllService.PeerService.InfoByRowId(1)
	if n != 1 || p.Alias != "Preset" || p.AbId != 1 {
		t.Fatalf("legacy preset behavior missing: n=%d peer=%+v", n, p)
	}
	service.AllService.AddressBookService.ApplyPresetToAddressBook(1, "101", "windows", "A", "Updated", "preset-host", "preset-user")
	ab := service.AllService.AddressBookService.InfoByUserIdAndIdAndCid(1, "101", 1)
	if ab.Alias != "Updated" {
		t.Fatal("legacy preset must update existing alias")
	}
}

func TestDisabledSchemeRestoresLegacyEditDeleteAndIdUpdate(t *testing.T) {
	db := fixture(t)
	legacyMode(t)
	ab := &model.AddressBook{Id: "101", UserId: 1, CollectionId: 1, Tags: []byte("[]")}
	if err := service.AllService.AddressBookService.Create(ab); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf(`{"row_id":%d,"id":"101","user_id":1,"collection_id":1,"alias":"Edited","hostname":"Book host","username":"Book user"}`, ab.RowId)
	success(t, call(t, 1, (&admin.AddressBook{}).Update, "POST", "/edit", body))
	p := service.AllService.PeerService.InfoByRowId(1)
	if p.Alias != "Edited" || p.AbId != 1 || p.Hostname != "Book host" {
		t.Fatal("legacy book edit no longer updates peer")
	}
	success(t, call(t, 1, (&admin.AddressBook{}).Delete, "POST", "/delete", fmt.Sprintf(`{"row_id":%d}`, ab.RowId)))
	db.First(p, 1)
	if p.Alias != "" || p.AbId != 0 {
		t.Fatal("legacy delete did not clear peer alias")
	}
	success(t, call(t, 1, (&admin.Peer{}).Update, "POST", "/peer-edit", `{"row_id":1,"id":"legacy-renamed"}`))
	if service.AllService.PeerService.InfoByRowId(1).Id != "legacy-renamed" {
		t.Fatal("legacy peer update changed")
	}
}

func TestDisabledSchemeIgnoresNewCopyInputs(t *testing.T) {
	db := fixture(t)
	db.Model(&model.Peer{}).Where("row_id=1").Update("connection_address", "device-default")
	legacyMode(t)
	success(t, call(t, 1, (&admin.AddressBook{}).BatchCreateFromPeers, "POST", "/copy", `{"peer_ids":[1],"user_id":1,"collection_id":1,"collection_ids":[1,2],"connection_address":"override","alias":"new-scheme-alias"}`))
	var books []model.AddressBook
	db.Find(&books)
	if len(books) != 1 || books[0].Id != "101" || books[0].CollectionId != 1 || books[0].PeerRowId != 0 || books[0].Alias != "" {
		t.Fatalf("new inputs affected legacy copy: %+v", books)
	}
}
