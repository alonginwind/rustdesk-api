package api

import "github.com/lejianwen/rustdesk-api/v2/model"

/*
	pub enum UserStatus {
	    Disabled = 0,
	    Normal = 1,
	    Unverified = -1,
	}
*/

/*
UserPayload
String name = ”;
String email = ”;
String note = ”;
UserStatus status;
bool isAdmin = false;
*/
type UserPayload struct {
	Name    string                 `json:"name"`
	Email   string                 `json:"email"`
	Note    string                 `json:"note"`
	IsAdmin *bool                  `json:"is_admin"`
	Status  int                    `json:"status"`
	Info    map[string]interface{} `json:"info"`
}

func (up *UserPayload) FromUser(user *model.User) *UserPayload {
	up.Name = user.Username
	up.Email = user.Email
	up.IsAdmin = user.IsAdmin
	up.Status = int(user.Status)
	up.Info = map[string]interface{}{}
	return up
}

/*
	class HttpType {
	  static const kAuthReqTypeAccount = "account";
	  static const kAuthReqTypeMobile = "mobile";
	  static const kAuthReqTypeSMSCode = "sms_code";
	  static const kAuthReqTypeEmailCode = "email_code";
	  static const kAuthReqTypeTfaCode = "tfa_code";

	  static const kAuthResTypeToken = "access_token";
	  static const kAuthResTypeEmailCheck = "email_check";
	  static const kAuthResTypeTfaCheck = "tfa_check";
	}
*/
type LoginRes struct {
	Type             string                `json:"type"`
	AccessToken      string                `json:"access_token"`
	User             UserPayload           `json:"user"`
	Secret           string                `json:"secret,omitempty"`
	TfaType          string            	   `json:"tfa_type,omitempty"`
	// 主控端策略，登录时随响应下发；字段名与被控端心跳的 strategy/modified_at 隔离，
	// 避免客户端共用同一个本地时间戳（strategy_timestamp）导致两条通道互相覆盖震荡
	// 客户端比对本地 client_strategy_timestamp，不一致时应用 config_options 并更新时间戳
	ClientStrategy   ClientStrategyPayload `json:"client_strategy"`
	ClientModifiedAt int64                 `json:"client_modified_at"`
}

// ClientStrategyPayload 登录响应中的主控端策略载荷
// 结构与被控端策略一致（config_options 键值映射），但承载通道与本地时间戳均独立
type ClientStrategyPayload struct {
	ConfigOptions map[string]string `json:"config_options"`
}

// CurrentUserRes /api/currentUser（含 /api/user/info）响应
// 嵌入 UserPayload 使字段平铺到顶层，与旧响应形态完全兼容，旧客户端忽略新增键
// 客户端每次启动经 refreshCurrentUser 调用本接口（token 非空时），
// 使 token 自动登录（免重新输密码）的主控端也能拿到最新策略
// 策略字段语义同 LoginRes：字段名与被控端心跳隔离，客户端比对 client_strategy_timestamp
type CurrentUserRes struct {
	UserPayload
	ClientStrategy   ClientStrategyPayload `json:"client_strategy"`
	ClientModifiedAt int64                 `json:"client_modified_at"`
}
