package model

type OperationLog struct {
	IdModel
	UserId   uint   `json:"user_id" gorm:"default:0;not null;index"`
	Username string `json:"username" gorm:"default:'';not null;"`
	Op       string `json:"op" gorm:"default:'';not null;"`       // create, update, delete, batch_delete
	Resource string `json:"resource" gorm:"default:'';not null;"` // user, peer, address_book ...
	Detail   string `json:"detail" gorm:"type:text;"`             // 请求体摘要
	Ip       string `json:"ip" gorm:"default:'';not null;"`
	Path     string `json:"path" gorm:"default:'';not null;"`     // 请求路径
	TimeModel
}

type OperationLogList struct {
	OperationLogs []*OperationLog `json:"list"`
	Pagination
}
