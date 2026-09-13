package service

import (
	"github.com/lejianwen/rustdesk-api/v2/config"
	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuditService struct {
}

func (as *AuditService) AuditConnList(page, pageSize uint, where func(tx *gorm.DB)) (res *model.AuditConnList) {
	res = &model.AuditConnList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.AuditConn{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.AuditConns)
	return
}

// CreateAuditConnIfNonceUnique creates the record only when no existing record
// carries the same nonce. An empty nonce always inserts (backward compatibility
// with clients that don't send one). The check and the insert happen inside a
// single transaction so concurrent file-audit posts from the client cannot
// sneak a duplicate between the two statements.
func (as *AuditService) CreateAuditConnIfNonceUnique(ac *model.AuditConn) error {
	if ac.Nonce == "" {
		return DB.Create(ac).Error
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		exists, err := as.connNonceExists(tx, ac.Nonce)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		return tx.Create(ac).Error
	})
}

// connNonceExists checks whether a record with the given nonce already exists.
// For MySQL/PostgreSQL the SELECT runs with FOR UPDATE so no other transaction
// can insert the same nonce between this check and the subsequent Create.
// SQLite relies on the IMMEDIATE transaction instead (FOR UPDATE is not
// supported).
func (as *AuditService) connNonceExists(tx *gorm.DB, nonce string) (bool, error) {
	var count int64
	q := tx.Model(&model.AuditConn{}).Where("nonce = ?", nonce)
	if Config.Gorm.Type != config.TypeSqlite {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
func (as *AuditService) DeleteAuditConn(u *model.AuditConn) error {
	return DB.Delete(u).Error
}

// Update 更新
func (as *AuditService) UpdateAuditConn(u *model.AuditConn) error {
	return DB.Model(u).Updates(u).Error
}

// InfoByPeerIdAndConnId returns the record a post belongs to. The controlled
// side numbers its connections from 1 in every process, so the pair also
// matches records of earlier runs: the live connection is the newest record
// still open, and that same record once it has been closed.
func (as *AuditService) InfoByPeerIdAndConnId(peerId string, connId int64) (res *model.AuditConn) {
	res = &model.AuditConn{}
	DB.Where("peer_id = ? and conn_id = ? and close_time = 0", peerId, connId).Order("id desc").First(res)
	if res.Id == 0 {
		DB.Where("peer_id = ? and conn_id = ?", peerId, connId).Order("id desc").First(res)
	}
	return
}

// ConnInfoByNonce finds a record by the tag the client gives each of its posts.
// A client re-sends a post it is unsure got stored, and the nonce is what tells
// that resend from a new connection. Posts predating the nonce send "", which
// is never looked up, so nothing is deduped for them.
func (as *AuditService) ConnInfoByNonce(nonce string) (res *model.AuditConn) {
	res = &model.AuditConn{}
	if nonce != "" {
		DB.Where("nonce = ?", nonce).First(res)
	}
	return
}

// ConnInfoById
func (as *AuditService) ConnInfoById(id uint) (res *model.AuditConn) {
	res = &model.AuditConn{}
	DB.Where("id = ?", id).First(res)
	return
}

// FileInfoById
func (as *AuditService) FileInfoById(id uint) (res *model.AuditFile) {
	res = &model.AuditFile{}
	DB.Where("id = ?", id).First(res)
	return
}

func (as *AuditService) AuditFileList(page, pageSize uint, where func(tx *gorm.DB)) (res *model.AuditFileList) {
	res = &model.AuditFileList{}
	res.Page = int64(page)
	res.PageSize = int64(pageSize)
	tx := DB.Model(&model.AuditFile{})
	if where != nil {
		where(tx)
	}
	tx.Count(&res.Total)
	tx.Scopes(Paginate(page, pageSize))
	tx.Find(&res.AuditFiles)
	return
}

// CreateAuditFileIfNonceUnique see CreateAuditConnIfNonceUnique
func (as *AuditService) CreateAuditFileIfNonceUnique(af *model.AuditFile) error {
	if af.Nonce == "" {
		return DB.Create(af).Error
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		exists, err := as.fileNonceExists(tx, af.Nonce)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		return tx.Create(af).Error
	})
}

// fileNonceExists see connNonceExists
func (as *AuditService) fileNonceExists(tx *gorm.DB, nonce string) (bool, error) {
	var count int64
	q := tx.Model(&model.AuditFile{}).Where("nonce = ?", nonce)
	if Config.Gorm.Type != config.TypeSqlite {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// FileInfoByNonce see ConnInfoByNonce
func (as *AuditService) FileInfoByNonce(nonce string) (res *model.AuditFile) {
	res = &model.AuditFile{}
	if nonce != "" {
		DB.Where("nonce = ?", nonce).First(res)
	}
	return
}

func (as *AuditService) DeleteAuditFile(u *model.AuditFile) error {
	return DB.Delete(u).Error
}

// Update 更新
func (as *AuditService) UpdateAuditFile(u *model.AuditFile) error {
	return DB.Model(u).Updates(u).Error
}

func (as *AuditService) BatchDeleteAuditConn(ids []uint) error {
	return DB.Where("id in (?)", ids).Delete(&model.AuditConn{}).Error
}

func (as *AuditService) BatchDeleteAuditFile(ids []uint) error {
	return DB.Where("id in (?)", ids).Delete(&model.AuditFile{}).Error
}
