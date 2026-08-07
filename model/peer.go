package model

type Peer struct {
	RowId          uint   `json:"row_id" gorm:"primaryKey;"`
	Id             string `json:"id"  gorm:"default:'';not null;index"`
	Cpu            string `json:"cpu"  gorm:"default:'';not null;"`
	Hostname       string `json:"hostname"  gorm:"default:'';not null;"`
	Memory         string `json:"memory"  gorm:"default:'';not null;"`
	Os             string `json:"os"  gorm:"default:'';not null;"`
	Username       string `json:"username"  gorm:"default:'';not null;"`
	Uuid           string `json:"uuid"  gorm:"default:'';not null;index"`
	Version        string `json:"version"  gorm:"default:'';not null;"`
	UserId         uint   `json:"user_id"  gorm:"default:0;not null;index"`
	User           *User  `json:"user,omitempty"`
	LastOnlineTime int64  `json:"last_online_time"  gorm:"default:0;not null;"`
	LastOnlineIp   string `json:"last_online_ip"  gorm:"default:'';not null;"`
	GroupId        uint   `json:"group_id"  gorm:"default:0;not null;index"`
	Alias          string `json:"alias" gorm:"default:'';not null;index"`
	// 预设地址簿信息，用于心跳下发到被控端
	PresetAbName  string `json:"preset_ab_name" gorm:"default:'';not null;"`
	PresetAbAlias string `json:"preset_ab_alias" gorm:"default:'';not null;"`
	PresetDevName string `json:"preset_dev_name" gorm:"default:'';not null;"`
	TimeModel
}

type PeerList struct {
	Peers []*Peer `json:"list"`
	Pagination
}
