package service

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/lejianwen/rustdesk-api/v2/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SQLite serializes writes; the mutex also makes duplicate checks atomic for a
// single API instance. Row locks serialize the same owner's writes on MySQL/PG.
var bookWriteMu sync.Mutex

func bookTransaction(ownerIds []uint, fn func(*gorm.DB) error) error {
	bookWriteMu.Lock()
	defer bookWriteMu.Unlock()
	sort.Slice(ownerIds, func(i, j int) bool { return ownerIds[i] < ownerIds[j] })
	return DB.Transaction(func(tx *gorm.DB) error {
		previous := uint(0)
		for _, id := range ownerIds {
			if id == 0 {
				return errors.New("ParamsError")
			}
			if id == previous {
				continue
			}
			previous = id
			var owner model.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&owner, id).Error; err != nil {
				return errors.New("ItemNotFound")
			}
		}
		return fn(tx)
	})
}

func validateBook(tx *gorm.DB, ab *model.AddressBook) error {
	if strings.TrimSpace(ab.Id) == "" || len(ab.Id) > 2048 || strings.ContainsAny(ab.Id, "\r\n\x00") {
		return errors.New("ParamsError")
	}
	if ab.CollectionId > 0 {
		var n int64
		if err := tx.Model(&model.AddressBookCollection{}).Where("id = ? AND user_id = ?", ab.CollectionId, ab.UserId).Count(&n).Error; err != nil {
			return err
		}
		if n != 1 {
			return errors.New("NoAccess")
		}
	}
	return nil
}

// CreateEntries preserves old memberships and is all-or-nothing. Existing rows
// are never overwritten on retry (they may have a book-specific destination).
func (s *AddressBookService) CreateEntries(entries []*model.AddressBook) error {
	owners := make([]uint, 0, len(entries))
	for _, ab := range entries {
		owners = append(owners, ab.UserId)
	}
	return bookTransaction(owners, func(tx *gorm.DB) error {
		for _, ab := range entries {
			if err := validateBook(tx, ab); err != nil {
				return err
			}
		}
		for _, ab := range entries {
			q := tx.Model(&model.AddressBook{}).Where("user_id = ? AND collection_id = ?", ab.UserId, ab.CollectionId)
			if ab.PeerRowId > 0 {
				var peer model.Peer
				if err := tx.Select("id").First(&peer, ab.PeerRowId).Error; err != nil {
					return err
				}
				q = q.Where("id = ? OR peer_row_id = ? OR (peer_row_id = 0 AND id = ?)", ab.Id, ab.PeerRowId, peer.Id)
			} else {
				q = q.Where("id = ?", ab.Id)
			}
			var n int64
			if err := q.Count(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			if ab.PeerRowId == 0 {
				var peer model.Peer
				if err := tx.Select("row_id").Where("id = ?", ab.Id).Find(&peer).Error; err != nil {
					return err
				}
				ab.PeerRowId = peer.RowId
			}
			if len(ab.Tags) == 0 {
				ab.Tags = []byte("[]")
			}
			if err := tx.Create(ab).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AddressBookService) UpdateEntry(ab *model.AddressBook) error {
	var previous model.AddressBook
	if err := DB.Select("row_id", "user_id").First(&previous, ab.RowId).Error; err != nil {
		return err
	}
	return bookTransaction([]uint{previous.UserId, ab.UserId}, func(tx *gorm.DB) error {
		if err := validateBook(tx, ab); err != nil {
			return err
		}
		var old model.AddressBook
		if err := tx.First(&old, ab.RowId).Error; err != nil {
			return err
		}
		// Personal controllers enforce the authenticated owner. Admins retain
		// their existing ability to explicitly transfer a contact.
		if old.UserId != previous.UserId {
			return errors.New("NoAccess")
		}
		var n int64
		if err := tx.Model(&model.AddressBook{}).Where("user_id = ? AND collection_id = ? AND id = ? AND row_id <> ?", ab.UserId, ab.CollectionId, ab.Id, ab.RowId).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errors.New("ItemExists")
		}
		if ab.PeerRowId == 0 {
			ab.PeerRowId = old.PeerRowId
		}
		if ab.PeerRowId == 0 {
			var peer model.Peer
			if err := tx.Select("row_id").Where("id = ?", old.Id).Find(&peer).Error; err != nil {
				return err
			}
			ab.PeerRowId = peer.RowId
		}
		if len(ab.Tags) == 0 {
			ab.Tags = []byte("[]")
		}
		return tx.Model(ab).Select("*").Omit("created_at").Updates(ab).Error
	})
}
