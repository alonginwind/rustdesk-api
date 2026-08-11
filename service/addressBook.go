package service

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/model/custom_types"
	"gorm.io/gorm"
	"strings"
	"sync"
)

var (
	SharePeerIds   = make(map[string]map[string]struct{})
	SharePeerIdsMu sync.Mutex
)

// 所有连接类型
var ShareConnTypes = []string{"default_conn", "file_transfer", "terminal"}

type AddressBookService struct {
}

func (s *AddressBookService) Info(id string) *model.AddressBook {
	p := &model.AddressBook{}
	DB.Where("id = ?", id).First(p)
	return p
}

func (s *AddressBookService) InfoByUserIdAndId(userid uint, id string) *model.AddressBook {
	p := &model.AddressBook{}
	DB.Where("user_id = ? and id = ?", userid, id).First(p)
	return p
}

func (s *AddressBookService) InfoByUserIdAndIdAndCid(userid uint, id string, cid uint) *model.AddressBook {
	p := &model.AddressBook{}
	DB.Where("user_id = ? and id = ? and collection_id = ?", userid, id, cid).First(p)
	return p
}
func (s *AddressBookService) InfoByRowId(id uint) *model.AddressBook {
	p := &model.AddressBook{}
	DB.Where("row_id = ?", id).First(p)
	return p
}
func (s *AddressBookService) ListByUserId(userId, page, pageSize uint) (res *model.AddressBookList) {
	res = s.List(page, pageSize, func(tx *gorm.DB) {
		tx.Where("user_id = ?", userId)
	})
	return
}
func (s *AddressBookService) ListByUserIds(userIds []uint, page, pageSize uint) (res *model.AddressBookList) {
	res = s.List(page, pageSize, func(tx *gorm.DB) {
		tx.Where("user_id in (?)", userIds)
	})
	return
}

// AddAddressBook
func (s *AddressBookService) AddAddressBook(ab *model.AddressBook) error {
	return DB.Create(ab).Error
}

// UpdateAddressBook
func (s *AddressBookService) UpdateAddressBook(abs []*model.AddressBook, userId uint) error {
	//比较peers和数据库中的数据，如果peers中的数据在数据库中不存在，则添加，如果存在则更新，如果数据库中的数据在peers中不存在，则删除
	// 开始事务
	tx := DB.Begin()
	//1. 获取数据库中的数据
	var dbABs []*model.AddressBook
	tx.Where("user_id = ?", userId).Find(&dbABs)
	//2. 比较peers和数据库中的数据
	//2.1 获取peers中的id
	aBIds := make(map[string]*model.AddressBook)
	for _, ab := range abs {
		aBIds[ab.Id] = ab
	}
	//2.2 获取数据库中的id
	dbABIds := make(map[string]*model.AddressBook)
	for _, dbAb := range dbABs {
		dbABIds[dbAb.Id] = dbAb
	}
	//2.3 比较peers和数据库中的数据
	for id, ab := range aBIds {
		dbAB, ok := dbABIds[id]
		ab.UserId = userId
		if !ok {
			//添加
			if ab.Platform == "" || ab.Username == "" || ab.Hostname == "" {
				peer := AllService.PeerService.FindById(ab.Id)
				if peer.RowId != 0 {
					ab.Platform = AllService.AddressBookService.PlatformFromOs(peer.Os)
					ab.Username = peer.Username
					ab.Hostname = peer.Hostname
				}
			}
			tx.Create(ab)
		} else {
			//更新
			tx.Model(&model.AddressBook{}).Where("row_id = ?", dbAB.RowId).Updates(ab)
		}
	}
	//2.4 删除
	for id, dbAB := range dbABIds {
		_, ok := aBIds[id]
		if !ok {
			tx.Delete(dbAB)
		}
	}
	tx.Commit()
	return nil

}

func (s *AddressBookService) List(page, pageSize uint, where func(tx *gorm.DB)) (res *model.AddressBookList) {
	res = &model.AddressBookList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.AddressBook{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize)).
		Order("(CASE WHEN alias = '' THEN 0 ELSE 1 END) ASC").
		Order("collection_id ASC").
		Order("alias ASC")
	tx.Find(&res.AddressBooks)
	return
}

func (s *AddressBookService) FromPeer(peer *model.Peer) (a *model.AddressBook) {
	a = &model.AddressBook{}
	a.Id = peer.Id
	a.Username = peer.Username
	a.Hostname = peer.Hostname
	a.UserId = peer.UserId
	a.Platform = s.PlatformFromOs(peer.Os)
	return a
}

// Create 创建
func (s *AddressBookService) Create(u *model.AddressBook) error {
	res := DB.Create(u).Error
	return res
}

// 清理在其他全员集合中的旧条目
func (s *AddressBookService) CleanUp(peerId string, collectionId uint) {
	DB.Where("id = ? AND user_id = 1 AND collection_id != ?", peerId, collectionId).Delete(&model.AddressBook{})
}

func (s *AddressBookService) Delete(u *model.AddressBook) error {
	return DB.Delete(u).Error
}

// Update 更新
func (s *AddressBookService) Update(u *model.AddressBook) error {
	return DB.Model(u).Updates(u).Error
}

// UpdateByMap 更新
func (s *AddressBookService) UpdateByMap(u *model.AddressBook, data map[string]interface{}) error {
	return DB.Model(u).Updates(data).Error
}

// UpdateAll 更新
func (s *AddressBookService) UpdateAll(u *model.AddressBook) error {
	return DB.Model(u).Select("*").Omit("created_at").Updates(u).Error
}

// ShareByWebClient 分享
func (s *AddressBookService) ShareByWebClient(m *model.ShareRecord) error {
	s.AddShareByWebClientId(m.PeerId)
	m.ShareToken = uuid.New().String()
	return DB.Create(m).Error
}

// 添加ShareByWebClient PeerId，初始化所有连接类型
func (s *AddressBookService) AddShareByWebClientId(id string) {
	SharePeerIdsMu.Lock()
	defer SharePeerIdsMu.Unlock()
	connTypes := make(map[string]struct{})
	for _, ct := range ShareConnTypes {
		connTypes[ct] = struct{}{}
	}
	SharePeerIds[id] = connTypes
}

// ConsumeShareByWebClientId 查询并消费（删除）指定连接类型，返回是否存在
func (s *AddressBookService) ConsumeShareByWebClientId(id string, connType string) bool {
	SharePeerIdsMu.Lock()
	defer SharePeerIdsMu.Unlock()
	connTypes, ok := SharePeerIds[id]
	if !ok {
		return false
	}
	if _, exists := connTypes[connType]; !exists {
		return false
	}
	delete(connTypes, connType)
	if len(connTypes) == 0 {
		delete(SharePeerIds, id)
	}
	return true
}

// DeleteAllShareByWebClientId 清理所有连接类型
func (s *AddressBookService) DeleteAllShareByWebClientId(id string) {
	SharePeerIdsMu.Lock()
	defer SharePeerIdsMu.Unlock()
	delete(SharePeerIds, id)
}

// SharedPeer
func (s *AddressBookService) SharedPeer(shareToken string) *model.ShareRecord {
	m := &model.ShareRecord{}
	DB.Where("share_token = ?", shareToken).Last(m)
	return m
}

// PlatformFromOs
func (s *AddressBookService) PlatformFromOs(os string) string {
	if strings.Contains(os, "Android") || strings.Contains(os, "android") {
		return "Android"
	}
	if strings.Contains(os, "Windows") || strings.Contains(os, "windows") {
		return "Windows"
	}
	if strings.Contains(os, "Linux") || strings.Contains(os, "linux") {
		return "Linux"
	}
	if strings.Contains(os, "mac") || strings.Contains(os, "Mac") {
		return "Mac OS"
	}
	return ""
}
func (s *AddressBookService) ListByUserIdAndCollectionId(userId, cid, page, pageSize uint) (res *model.AddressBookList) {
	res = s.List(page, pageSize, func(tx *gorm.DB) {
		tx.Where("user_id = ? and collection_id = ?", userId, cid)
	})
	return
}
func (s *AddressBookService) ListCollection(page, pageSize uint, where func(tx *gorm.DB)) (res *model.AddressBookCollectionList) {
	res = &model.AddressBookCollectionList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.AddressBookCollection{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.AddressBookCollection)
	return
}
func (s *AddressBookService) ListCollectionByIds(ids []uint) (res []*model.AddressBookCollection) {
	DB.Where("id in ?", ids).Find(&res)
	return res
}

func (s *AddressBookService) ListCollectionByUserId(userId uint) (res *model.AddressBookCollectionList) {
	res = s.ListCollection(1, 100, func(tx *gorm.DB) {
		tx.Where("user_id = ?", userId)
	})
	return
}
func (s *AddressBookService) CollectionInfoById(id uint) *model.AddressBookCollection {
	p := &model.AddressBookCollection{}
	DB.Where("id = ?", id).First(p)
	return p
}

func (s *AddressBookService) CollectionReadRules(user *model.User) (res []*model.AddressBookCollectionRule) {
	// personalRules
	var personalRules []*model.AddressBookCollectionRule
	tx2 := DB.Model(&model.AddressBookCollectionRule{})
	tx2.Where("type = ? and to_id = ? and rule > 0", model.ShareAddressBookRuleTypePersonal, user.Id).Find(&personalRules)
	res = append(res, personalRules...)

	//group
	var groupRules []*model.AddressBookCollectionRule
	tx3 := DB.Model(&model.AddressBookCollectionRule{})
	tx3.Where("type = ? and to_id = ? and rule > 0", model.ShareAddressBookRuleTypeGroup, user.GroupId).Find(&groupRules)
	res = append(res, groupRules...)
	return
}

func (s *AddressBookService) UserMaxRule(user *model.User, uid, cid uint) int {
	// ismy?
	if user.Id == uid {
		return model.ShareAddressBookRuleRuleFullControl
	}
	max := 0
	personalRules := &model.AddressBookCollectionRule{}
	tx := DB.Model(personalRules)
	tx.Where("type = ? and collection_id = ? and to_id = ?", model.ShareAddressBookRuleTypePersonal, cid, user.Id).First(&personalRules)
	if personalRules.Id != 0 {
		max = personalRules.Rule
		if max == model.ShareAddressBookRuleRuleFullControl {
			return max
		}
	}

	groupRules := &model.AddressBookCollectionRule{}
	tx2 := DB.Model(groupRules)
	tx2.Where("type = ? and collection_id = ? and to_id = ?", model.ShareAddressBookRuleTypeGroup, cid, user.GroupId).First(&groupRules)
	if groupRules.Id != 0 {
		if groupRules.Rule > max {
			max = groupRules.Rule
		}
		if max == model.ShareAddressBookRuleRuleFullControl {
			return max
		}
	}
	return max
}

func (s *AddressBookService) CheckUserReadPrivilege(user *model.User, uid, cid uint) bool {
	return s.UserMaxRule(user, uid, cid) >= model.ShareAddressBookRuleRuleRead
}
func (s *AddressBookService) CheckUserWritePrivilege(user *model.User, uid, cid uint) bool {
	return s.UserMaxRule(user, uid, cid) >= model.ShareAddressBookRuleRuleReadWrite
}
func (s *AddressBookService) CheckUserFullControlPrivilege(user *model.User, uid, cid uint) bool {
	return s.UserMaxRule(user, uid, cid) >= model.ShareAddressBookRuleRuleFullControl
}

func (s *AddressBookService) CreateCollection(t *model.AddressBookCollection) error {
	return DB.Create(t).Error
}

func (s *AddressBookService) UpdateCollection(t *model.AddressBookCollection) error {
	return DB.Model(t).Updates(t).Error
}

func (s *AddressBookService) DeleteCollection(t *model.AddressBookCollection) error {
	//删除集合下的所有规则、地址簿，再删除集合
	tx := DB.Begin()
	tx.Where("collection_id = ?", t.Id).Delete(&model.AddressBookCollectionRule{})
	tx.Where("collection_id = ?", t.Id).Delete(&model.AddressBook{})
	tx.Delete(t)
	return tx.Commit().Error
}

func (s *AddressBookService) RuleInfoById(u uint) *model.AddressBookCollectionRule {
	p := &model.AddressBookCollectionRule{}
	DB.Where("id = ?", u).First(p)
	return p
}
func (s *AddressBookService) RulePersonalInfoByToIdAndCid(toid, cid uint) *model.AddressBookCollectionRule {
	return s.RuleInfoByToIdAndCid(model.ShareAddressBookRuleTypePersonal, toid, cid)
}
func (s *AddressBookService) RuleInfoByToIdAndCid(t int, toid, cid uint) *model.AddressBookCollectionRule {
	p := &model.AddressBookCollectionRule{}
	DB.Where("type = ? and to_id = ? and collection_id = ?", t, toid, cid).First(p)
	return p
}
func (s *AddressBookService) CreateRule(t *model.AddressBookCollectionRule) error {
	return DB.Create(t).Error
}

func (s *AddressBookService) ListRules(page uint, size uint, f func(tx *gorm.DB)) *model.AddressBookCollectionRuleList {
	res := &model.AddressBookCollectionRuleList{}
	res.Page = int64(page)
	res.PageSize = int64(size)
	tx := DB.Model(&model.AddressBookCollectionRule{})
	if f != nil {
		f(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, size))
	tx.Find(&res.AddressBookCollectionRule)
	return res
}

func (s *AddressBookService) UpdateRule(t *model.AddressBookCollectionRule) error {
	return DB.Model(t).Updates(t).Error
}

func (s *AddressBookService) DeleteRule(t *model.AddressBookCollectionRule) error {
	return DB.Delete(t).Error
}

// CheckCollectionOwner 检查Collection的所有者
func (s *AddressBookService) CheckCollectionOwner(uid uint, cid uint) bool {
	p := s.CollectionInfoById(cid)
	return p.UserId == uid
}

// ApplyPresetToAddressBook 根据客户端上报的预设值自动将设备添加到指定地址簿集合
// 参数：
//   peerId   - 设备ID
//   os       - 操作系统信息
//   abName   - preset-address-book-name，地址簿集合名称
//   abAlias  - preset-address-book-alias，地址簿中的别名
//   hostname - 主机名（客户端已将 preset-device-name 写入 hostname 字段）
//   username - 用户名（客户端已将 preset-device-username 写入 username 字段）
func (s *AddressBookService) ApplyPresetToAddressBook(peerId, os, abName, abAlias, hostname, username string) {
	if abName == "" {
		return
	}

	// 查找地址簿集合（admin用户 user_id=1 下）
	collection := &model.AddressBookCollection{}
	if DB.Where("name = ? AND user_id = 1", abName).First(collection).Error != nil {
		return
	}

	// 检查该集合下是否已有该设备的条目
	existing := &model.AddressBook{}
	if DB.Where("id = ? AND collection_id = ?", peerId, collection.Id).First(existing).Error == nil {
		// 已存在，仅用非空值更新，避免覆盖管理员手动设置的别名
		updates := map[string]interface{}{}
		if abAlias != "" {
			updates["alias"] = abAlias
		}
		if hostname != "" {
			updates["hostname"] = hostname
		}
		if len(updates) > 0 {
			DB.Model(existing).Updates(updates)
		}
		return
	}

	// 创建新条目
	platform := s.PlatformFromOs(os)
	ab := &model.AddressBook{
		Id:           peerId,
		Alias:        abAlias,
		Hostname:     hostname,
		Username:     username,
		Platform:     platform,
		Tags:         custom_types.AutoJson("[]"),
		UserId:       1,
		CollectionId: collection.Id,
	}
	if err := DB.Create(ab).Error; err != nil {
		// 创建失败时静默处理，不影响 sysinfo 主流程
		_ = err
	} else {
		// 删除该设备在其他全员集合中的旧条目
		s.CleanUp(peerId, collection.Id)
	}
}

// GetPresetValuesForPeer 从地址簿中查询设备的预设信息
// 返回: (collectionName, alias, hostname)
func (s *AddressBookService) GetPresetValuesForPeer(peerId string) (string, string, string) {
	if peerId == "" {
		return "", "", ""
	}
	// 查找地址簿条目（按设备ID查找全员地址簿）
	ab := &model.AddressBook{}
	if DB.Where("id = ? AND user_id = 1", peerId).First(ab).Error != nil {
		return "", "", ""
	}
	collectionName := ""
	if ab.CollectionId > 0 {
		collection := s.CollectionInfoById(ab.CollectionId)
		if collection.Id > 0 {
			collectionName = collection.Name
		}
	}
	return collectionName, ab.Alias, ab.Hostname
}

func (s *AddressBookService) BatchUpdateTags(abs []*model.AddressBook, tags []string) error {
	ids := make([]uint, 0)
	for _, ab := range abs {
		ids = append(ids, ab.RowId)
	}
	tagsv, _ := json.Marshal(tags)
	return DB.Model(&model.AddressBook{}).Where("row_id in ?", ids).Update("tags", tagsv).Error
}
