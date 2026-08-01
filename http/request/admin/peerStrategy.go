package admin

import (
	"encoding/json"

	"github.com/lejianwen/rustdesk-api/v2/model"
	"github.com/lejianwen/rustdesk-api/v2/model/custom_types"
)

// PeerStrategyForm 被控端策略表单
type PeerStrategyForm struct {
	Id            uint              `json:"id"`
	PeerId        string            `json:"peer_id" validate:"required"`
	ConfigOptions map[string]string `json:"config_options"`
}

// ToPeerStrategy 转换为 model.PeerStrategy
func (f *PeerStrategyForm) ToPeerStrategy() *model.PeerStrategy {
	configOptionsJson, _ := json.Marshal(f.ConfigOptions)
	ps := &model.PeerStrategy{}
	ps.Id = f.Id
	ps.PeerId = f.PeerId
	ps.ConfigOptions = custom_types.AutoJson(configOptionsJson)
	return ps
}

type PeerStrategyQuery struct {
	PageQuery
	PeerId string `json:"peer_id" form:"peer_id"`
}

// DefaultStrategyForm 默认策略表单
// 不需要 peer_id，因为默认策略的 peer_id 为空
type DefaultStrategyForm struct {
	ConfigOptions map[string]string `json:"config_options"`
}

func (f *DefaultStrategyForm) ToPeerStrategy() *model.PeerStrategy {
	configOptionsJson, _ := json.Marshal(f.ConfigOptions)
	ps := &model.PeerStrategy{}
	ps.PeerId = ""
	ps.ConfigOptions = custom_types.AutoJson(configOptionsJson)
	return ps
}
