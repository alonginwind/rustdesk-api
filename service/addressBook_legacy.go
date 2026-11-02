package service

// Original upstream behavior, used only when app.device-address-books is disabled.
import (
	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/model/custom_types"
)

func (s *AddressBookService) legacyAddAddressBook(ab *model.AddressBook) error {
	return DB.Create(ab).Error
}

func (s *AddressBookService) legacyUpdateAddressBook(abs []*model.AddressBook, userId uint) error {
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

func (s *AddressBookService) legacyFromPeer(peer *model.Peer) (a *model.AddressBook) {
	a = &model.AddressBook{}
	a.Id = peer.Id
	a.Username = peer.Username
	a.Hostname = peer.Hostname
	a.UserId = peer.UserId
	a.Platform = s.PlatformFromOs(peer.Os)
	return a
}

func (s *AddressBookService) legacyCreate(u *model.AddressBook) error {
	res := DB.Create(u).Error
	return res
}

func (s *AddressBookService) legacyUpdateAll(u *model.AddressBook) error {
	return DB.Model(u).Select("*").Omit("created_at", "peer_row_id").Updates(u).Error
}

func (s *AddressBookService) legacyApplyPresetToAddressBook(peerRowId uint, peerId, os, abName, abAlias, hostname, username string) {
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
		if username != "" {
			updates["username"] = username
		}
		if len(updates) > 0 {
			DB.Model(existing).Updates(updates)
		}
		if peerRowId > 0 {
			AllService.PeerService.UpdateAlias(peerRowId, collection.Id, abAlias, username, hostname)
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
		if peerRowId > 0 {
			AllService.PeerService.UpdateAlias(peerRowId, collection.Id, abAlias, username, hostname)
		}
	}
}

func (s *AddressBookService) CleanUp(peerId string, collectionId uint) {
	DB.Where("id = ? AND user_id = 1 AND collection_id != ?", peerId, collectionId).Delete(&model.AddressBook{})
}
