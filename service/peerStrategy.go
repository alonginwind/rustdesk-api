package service

import (
	"encoding/json"
	"time"

	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PeerStrategyService struct{}

// FindByPeerId 根据设备ID查找策略（不含默认策略）
func (pss *PeerStrategyService) FindByPeerId(peerId string) *model.PeerStrategy {
	ps := &model.PeerStrategy{}
	DB.Where("peer_id = ? AND peer_id != ''", peerId).First(ps)
	return ps
}

// FindDefault 查找全局默认策略（peer_id 为空的记录）
func (pss *PeerStrategyService) FindDefault() *model.PeerStrategy {
	ps := &model.PeerStrategy{}
	DB.Where("peer_id = ''").First(ps)
	return ps
}

// CreateOrUpdateDefault 创建或更新默认策略
// 使用 OnConflict 实现幂等 upsert，避免并发竞态
func (pss *PeerStrategyService) CreateOrUpdateDefault(ps *model.PeerStrategy) error {
	ps.PeerId = "" // 强制保证默认策略的 peer_id 为空
	ps.ModifiedAt = time.Now().UnixMilli()
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "peer_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"config_options",
			"modified_at",
			"updated_at",
		}),
	}).Create(ps).Error
}

// Info 根据主键查找
func (pss *PeerStrategyService) Info(id uint) *model.PeerStrategy {
	ps := &model.PeerStrategy{}
	DB.Where("id = ?", id).First(ps)
	return ps
}

// List 分页列表
func (pss *PeerStrategyService) List(page, pageSize uint, where func(tx *gorm.DB)) (res *model.PeerStrategyList) {
	res = &model.PeerStrategyList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.PeerStrategy{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.PeerStrategies)
	// 批量查询关联的 peer 别名
	if len(res.PeerStrategies) > 0 {
		peerIds := make([]string, 0, len(res.PeerStrategies))
		for _, ps := range res.PeerStrategies {
			if ps.PeerId != "" {
				peerIds = append(peerIds, ps.PeerId)
			}
		}
		if len(peerIds) > 0 {
			peers := make([]*model.Peer, 0)
			DB.Select("id, alias").Where("id IN ?", peerIds).Find(&peers)
			// 构建 peer_id -> alias 映射
			aliasMap := make(map[string]string)
			for _, p := range peers {
				aliasMap[p.Id] = p.Alias
			}
			// 填充别名
			for _, ps := range res.PeerStrategies {
				if alias, ok := aliasMap[ps.PeerId]; ok {
					ps.PeerAlias = alias
				}
			}
		}
	}
	return
}

// Create 创建策略
func (pss *PeerStrategyService) Create(ps *model.PeerStrategy) error {
	ps.ModifiedAt = time.Now().UnixMilli()
	return DB.Create(ps).Error
}

// Update 更新策略，同时刷新 ModifiedAt
func (pss *PeerStrategyService) Update(ps *model.PeerStrategy) error {
	ps.ModifiedAt = time.Now().UnixMilli()
	return DB.Save(ps).Error
}

// Delete 删除策略
func (pss *PeerStrategyService) Delete(ps *model.PeerStrategy) error {
	return DB.Delete(ps).Error
}

// GetConfigOptions 将 ConfigOptions (AutoJson) 解析为 map[string]string
// 如果解析失败或为空，返回空 map
func (pss *PeerStrategyService) GetConfigOptions(ps *model.PeerStrategy) map[string]string {
	configOptions := map[string]string{}
	if ps == nil || len(ps.ConfigOptions) == 0 {
		return configOptions
	}
	if err := json.Unmarshal([]byte(ps.ConfigOptions), &configOptions); err != nil {
		// 解析失败（可能是空数组 []），返回空 map
		return map[string]string{}
	}
	if configOptions == nil {
		configOptions = map[string]string{}
	}
	return configOptions
}

// GetMergedConfigOptions 合并设备策略和默认策略的配置
// 优先级：设备策略 > 默认策略
// 如果设备策略没有某个配置项，则使用默认策略的配置
func (pss *PeerStrategyService) GetMergedConfigOptions(deviceStrategy, defaultStrategy *model.PeerStrategy) map[string]string {
	// 先从默认策略开始
	merged := pss.GetConfigOptions(defaultStrategy)
	// 用设备策略覆盖
	deviceOptions := pss.GetConfigOptions(deviceStrategy)
	for k, v := range deviceOptions {
		merged[k] = v
	}
	return merged
}

// GetEffectiveModifiedAt 获取有效的 modified_at（取设备策略和默认策略的最大值）
func (pss *PeerStrategyService) GetEffectiveModifiedAt(deviceStrategy, defaultStrategy *model.PeerStrategy) int64 {
	deviceModifiedAt := int64(0)
	if deviceStrategy != nil && deviceStrategy.Id > 0 {
		deviceModifiedAt = deviceStrategy.ModifiedAt
	}
	defaultModifiedAt := int64(0)
	if defaultStrategy != nil && defaultStrategy.Id > 0 {
		defaultModifiedAt = defaultStrategy.ModifiedAt
	}
	if deviceModifiedAt > defaultModifiedAt {
		return deviceModifiedAt
	}
	return defaultModifiedAt
}
