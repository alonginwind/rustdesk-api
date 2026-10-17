package service

import (
	"sync"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/gorm"
)

type heartbeatEntry struct {
	Time int64
	Ip   string
}

var (
	heartbeatMu     sync.Mutex
	heartbeatBuffer = make(map[uint]*heartbeatEntry)
)

// StartHeartbeatFlusher 启动后台协程，定时将心跳缓冲批量刷入数据库
func (ps *PeerService) StartHeartbeatFlusher() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			ps.flushHeartbeatBuffer()
		}
	}()
}

// BufferHeartbeat 心跳在线信息先写内存，后台每10s批量刷入数据库
func (ps *PeerService) BufferHeartbeat(rowId uint, t int64, ip string) {
	heartbeatMu.Lock()
	defer heartbeatMu.Unlock()
	heartbeatBuffer[rowId] = &heartbeatEntry{Time: t, Ip: ip}
}

func (ps *PeerService) flushHeartbeatBuffer() {
	heartbeatMu.Lock()
	if len(heartbeatBuffer) == 0 {
		heartbeatMu.Unlock()
		return
	}
	buf := heartbeatBuffer
	heartbeatBuffer = make(map[uint]*heartbeatEntry)
	heartbeatMu.Unlock()

	tx := DB.Begin()
	for rowId, e := range buf {
		if err := tx.Model(&model.Peer{}).Where("row_id = ?", rowId).Updates(map[string]interface{}{
			"last_online_time": e.Time,
			"last_online_ip":   e.Ip,
		}).Error; err != nil {
			tx.Rollback()
			Logger.Warnf("flush heartbeat buffer failed at peer %d: %v", rowId, err)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		Logger.Warnf("flush heartbeat buffer commit failed: %v", err)
	}
}

type PeerService struct {
}

// FindById 根据id查找
func (ps *PeerService) FindById(id string) *model.Peer {
	p := &model.Peer{}
	DB.Where("id = ?", id).First(p)
	return p
}
func (ps *PeerService) FindByUuid(uuid string) *model.Peer {
	p := &model.Peer{}
	DB.Where("uuid = ?", uuid).First(p)
	return p
}
func (ps *PeerService) InfoByRowId(id uint) *model.Peer {
	p := &model.Peer{}
	DB.Where("row_id = ?", id).First(p)
	return p
}

// FindByUserIdAndUuid 根据用户id和uuid查找peer
func (ps *PeerService) FindByUserIdAndUuid(uuid string, userId uint) *model.Peer {
	p := &model.Peer{}
	DB.Where("uuid = ? and user_id = ?", uuid, userId).First(p)
	return p
}

// UuidBindUserId 绑定用户id
func (ps *PeerService) UuidBindUserId(deviceId string, uuid string, userId uint) {
	peer := ps.FindByUuid(uuid)
	// 如果存在则更新
	if peer.RowId > 0 {
		peer.UserId = userId
		ps.Update(peer)
	} else {
		// 不存在则创建
		/*if deviceId != "" {
			DB.Create(&model.Peer{
				Id:     deviceId,
				Uuid:   uuid,
				UserId: userId,
			})
		}*/
	}
}

// UuidUnbindUserId 解绑用户id, 用于用户注销
func (ps *PeerService) UuidUnbindUserId(uuid string, userId uint) {
	peer := ps.FindByUserIdAndUuid(uuid, userId)
	if peer.RowId > 0 {
		DB.Model(peer).Update("user_id", 0)
	}
}

// EraseUserId 清除用户id, 用于用户删除
func (ps *PeerService) EraseUserId(userId uint) error {
	return DB.Model(&model.Peer{}).Where("user_id = ?", userId).Update("user_id", 0).Error
}

// ListByUserIds 根据用户id取列表
func (ps *PeerService) ListByUserIds(userIds []uint, page, pageSize uint) (res *model.PeerList) {
	res = &model.PeerList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.Peer{})
	tx.Where("user_id in (?)", userIds)
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.Peers)
	return
}

func (ps *PeerService) List(page, pageSize uint, orderBy string, where func(tx *gorm.DB)) (res *model.PeerList) {
	res = &model.PeerList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.Peer{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)

	orderClause := "created_at DESC"
	if orderBy == "alias" {
		orderClause = "alias ASC"
	} else if orderBy == "time" {
		orderClause = "created_at DESC"
	}
	tx = tx.Order(orderClause)
	tx = tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.Peers)
	return
}

// ListFilterByUserId 根据用户id过滤Peer列表
func (ps *PeerService) ListFilterByUserId(page, pageSize uint, where func(tx *gorm.DB), userId uint) (res *model.PeerList) {
	userWhere := func(tx *gorm.DB) {
		tx.Where("user_id = ?", userId)
		// 如果还有额外的筛选条件，执行它
		if where != nil {
			where(tx)
		}
	}
	return ps.List(page, pageSize, "alias", userWhere)
}

// Create 创建
func (ps *PeerService) Create(u *model.Peer) error {
	res := DB.Create(u).Error
	return res
}

// Delete 删除, 同时也应该删除token
func (ps *PeerService) Delete(u *model.Peer) error {
	uuid := u.Uuid
	err := DB.Delete(u).Error
	if err != nil {
		return err
	}
	// 删除token
	return AllService.UserService.FlushTokenByUuid(uuid)
}

// GetUuidListByIDs 根据ids获取uuid列表
func (ps *PeerService) GetUuidListByIDs(ids []uint) ([]string, error) {
	var uuids []string
	err := DB.Model(&model.Peer{}).
		Where("row_id in (?)", ids).
		Pluck("uuid", &uuids).Error
	//过滤uuids中的空字符串
	var newUuids []string
	for _, uuid := range uuids {
		if uuid != "" {
			newUuids = append(newUuids, uuid)
		}
	}
	return newUuids, err
}

// BatchDelete 批量删除, 同时也应该删除token
func (ps *PeerService) BatchDelete(ids []uint) error {
	uuids, err := ps.GetUuidListByIDs(ids)
	err = DB.Where("row_id in (?)", ids).Delete(&model.Peer{}).Error
	if err != nil {
		return err
	}
	// 删除token
	return AllService.UserService.FlushTokenByUuids(uuids)
}

// Update 更新
func (ps *PeerService) Update(u *model.Peer) error {
	return DB.Model(u).Updates(u).Error
}

// ClearAlias 置空设备别名（支持空值写入，struct Updates 无法做到）
func (ps *PeerService) ClearAlias(rowId uint) error {
	return DB.Model(&model.Peer{}).Where("row_id = ?", rowId).Updates(map[string]interface{}{
		"ab_id": 0,
		"alias": "",
	}).Error
}

// UpdateAlias 更新设备别名（支持空值写入）
func (ps *PeerService) UpdateAlias(rowId, abId uint, alias, username, hostname string) error {
	return DB.Model(&model.Peer{}).Where("row_id = ?", rowId).Updates(map[string]interface{}{
		"ab_id": abId,
		"alias": alias,
		"username": username,
		"hostname": hostname,
	}).Error
}

// UpdatePresets 仅更新预设地址簿字段（支持空值写入）
func (ps *PeerService) UpdatePresets(rowId, abId uint, presetAbAlias, presetDevUsrName, presetDevName string, presetModifiedAt int64) error {
	return DB.Model(&model.Peer{}).Where("row_id = ?", rowId).Updates(map[string]interface{}{
		"preset_ab_id":  abId,
		"preset_ab_alias": presetAbAlias,
		"preset_dev_usr_name": presetDevUsrName,
		"preset_dev_name": presetDevName,
		"preset_modified_at": presetModifiedAt,
	}).Error
}
