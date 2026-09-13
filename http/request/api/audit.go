package api

import (
	"encoding/json"
	"github.com/lejianwen/rustdesk-api/v2/global"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"strconv"
	"strings"
)

type AuditConnForm struct {
	Action       string   `json:"action"`
	ConnId       int64    `json:"conn_id"`
	Id           string   `json:"id"`
	Peer         []string `json:"peer"`
	Ip           string   `json:"ip"`
	SessionId    float64  `json:"session_id"`
	Type         int      `json:"type"`
	Uuid         string   `json:"uuid"`
	Nonce        string   `json:"nonce"`
	ConnAuditRef string   `json:"conn_audit_ref"`
	PrimaryAuth  int      `json:"primary_auth"`
	TwoFactor    int      `json:"two_factor"`
}

func (a *AuditConnForm) ToAuditConn() *model.AuditConn {
	fp := ""
	fn := ""
	if len(a.Peer) >= 1 {
		fp = a.Peer[0]
		if len(a.Peer) == 2 {
			fn = a.Peer[1]
		}
	}
	ssid := strconv.FormatFloat(a.SessionId, 'f', -1, 64)
	return &model.AuditConn{
		Action:       a.Action,
		ConnId:       a.ConnId,
		PeerId:       a.Id,
		FromPeer:     fp,
		FromName:     fn,
		Ip:           a.Ip,
		SessionId:    ssid,
		Type:         a.Type,
		Uuid:         a.Uuid,
		Nonce:        a.Nonce,
		ConnAuditRef: a.ConnAuditRef,
		PrimaryAuth:  a.PrimaryAuth,
		TwoFactor:    a.TwoFactor,
	}
}

type AuditFileForm struct {
	Id     string `json:"id"`
	Info   string `json:"info"`
	IsFile bool   `json:"is_file"`
	Path   string `json:"path"`
	PeerId string `json:"peer_id"`
	Type   int    `json:"type"`
	Uuid   string `json:"uuid"`
	Nonce  string `json:"nonce"`
	ConnId int64  `json:"conn_id"`
}
type AuditFileInfo struct {
	Ip   string `json:"ip"`
	Name string `json:"name"`
	Num  int    `json:"num"`
}

func (a *AuditFileForm) ToAuditFile() *model.AuditFile {
	fi := &AuditFileInfo{}
	err := json.Unmarshal([]byte(a.Info), fi)
	if err != nil {
		global.Logger.Warn("ToAuditFile", err)
	}
	info, named := a.fillFileNames()

	return &model.AuditFile{
		PeerId:   a.Id,
		Info:     info,
		IsFile:   a.IsFile && !named,
		FromPeer: a.PeerId,
		Path:     a.Path,
		Type:     a.Type,
		Uuid:     a.Uuid,
		Nonce:    a.Nonce,
		ConnId:   a.ConnId,
		FromName: fi.Name,
		Ip:       fi.Ip,
		Num:      fi.Num,
	}
}

// fillFileNames recovers the file names a transfer left out. A single-file job
// carries none: the controlled side knows only the file's path, so it sends the
// name empty and flags the record as a plain file. Once the name is back the
// flag is dropped, since that flag is what makes the admin list show the size.
func (a *AuditFileForm) fillFileNames() (info string, named bool) {
	name := auditFileName(a.Path)
	if a.Info == "" || name == "" {
		return a.Info, false
	}
	x := map[string]interface{}{}
	if err := json.Unmarshal([]byte(a.Info), &x); err != nil {
		return a.Info, false
	}
	files, ok := x["files"].([]interface{})
	if !ok {
		return a.Info, false
	}
	for _, f := range files {
		entry, ok := f.([]interface{})
		if !ok || len(entry) == 0 {
			continue
		}
		if n, ok := entry[0].(string); ok && n == "" {
			entry[0] = name
			named = true
		}
	}
	if !named {
		return a.Info, false
	}
	b, err := json.Marshal(x)
	if err != nil {
		return a.Info, false
	}
	return string(b), true
}

// auditFileName is the last segment of a path, which the transferring device
// may have written with either separator.
func auditFileName(path string) string {
	if i := strings.LastIndexAny(path, `/\`); i >= 0 {
		return path[i+1:]
	}
	return path
}
