package model

import "github.com/lejianwen/rustdesk-api/v2/model/custom_types"

// PeerStrategy 被控端策略配置
// 对应 RustDesk 客户端 sync.rs 中心跳响应的 strategy 字段
// 通过心跳接口下发给被控端，被控端收到后会应用 config_options 到本地配置
type PeerStrategy struct {
	IdModel
	PeerId        string                `json:"peer_id" gorm:"default:'';not null;uniqueIndex"` // 关联 peer.id
	ConfigOptions custom_types.AutoJson `json:"config_options" gorm:"not null;" swaggertype:"object,string"` // HashMap<String,String> 的 JSON
	ModifiedAt    int64                 `json:"modified_at" gorm:"default:0;not null;"` // 策略版本时间戳，用于变更检测
	TimeModel
	PeerAlias string `json:"peer_alias" gorm:"-"` // 关联的 peer 别名，非数据库字段
}

type PeerStrategyList struct {
	PeerStrategies []*PeerStrategy `json:"list"`
	Pagination
}
