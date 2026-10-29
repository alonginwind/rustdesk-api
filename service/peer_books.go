package service

import (
	"errors"
	"strings"

	"github.com/lejianwen/rustdesk-api/v2/model"
)

// DecorateBooks resolves legacy ID contacts and explicit device associations in
// one query. A caller must supply a viewer unless it is an administrator route.
func (ps *PeerService) DecorateBooks(peers []*model.Peer, viewer *model.User, administrator bool) error {
	if len(peers) == 0 {
		return nil
	}
	rows, ids := make([]uint, 0, len(peers)), make([]string, 0, len(peers))
	for _, p := range peers {
		rows = append(rows, p.RowId)
		ids = append(ids, p.Id)
		p.BookMemberships = make([]model.PeerBookMembership, 0)
		p.ConnectionAddress = p.Address()
	}
	tx := DB.Table("address_books AS b").Select("b.row_id, b.peer_row_id, b.user_id, b.collection_id, COALESCE(c.name, '') AS collection_name, b.alias, b.id AS connection_address").Joins("LEFT JOIN address_book_collections AS c ON c.id = b.collection_id").Where("(b.peer_row_id IN ? OR (b.peer_row_id = 0 AND b.id IN ?))", rows, ids)
	if !administrator {
		if viewer == nil || viewer.Id == 0 {
			return errors.New("NoAccess")
		}
		// Shared contact visibility follows the same read rules as native books.
		shared := DB.Model(&model.AddressBookCollectionRule{}).Select("collection_id").Where("rule >= ? AND ((type = ? AND to_id = ?) OR (type = ? AND to_id = ?))", model.ShareAddressBookRuleRuleRead, model.ShareAddressBookRuleTypePersonal, viewer.Id, model.ShareAddressBookRuleTypeGroup, viewer.GroupId)
		tx = tx.Where("b.user_id = ? OR (b.collection_id > 0 AND b.collection_id IN (?))", viewer.Id, shared)
	}
	var entries []model.PeerBookMembership
	if err := tx.Order("b.user_id, b.collection_id, b.row_id").Scan(&entries).Error; err != nil {
		return err
	}
	for _, e := range entries {
		for _, p := range peers {
			if e.PeerRowId == p.RowId || (e.PeerRowId == 0 && e.ConnectionAddress == p.Id) {
				p.BookMemberships = append(p.BookMemberships, e)
			}
		}
	}
	return nil
}

// SetConnectionAddress only changes the template used by future book copies.
func (ps *PeerService) SetConnectionAddress(rowId uint, address string, owner *uint) error {
	address = strings.TrimSpace(address)
	if len(address) > 2048 || strings.ContainsAny(address, "\r\n\x00") {
		return errors.New("ParamsError")
	}
	tx := DB.Model(&model.Peer{}).Where("row_id = ?", rowId)
	if owner != nil {
		tx = tx.Where("user_id = ?", *owner)
	}
	res := tx.Update("connection_address", address)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		if err := tx.Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return errors.New("ItemNotFound")
		}
	}
	return nil
}

// CopyPeersToBooks is atomic across all requested peers and books. A repeated
// request keeps existing local aliases, credentials and destinations intact.
func (s *AddressBookService) CopyPeersToBooks(peerIds, collectionIds []uint, userId uint, onlyOwned bool, address, alias *string, tags []byte) error {
	if userId == 0 || len(peerIds) == 0 || len(collectionIds) == 0 {
		return errors.New("ParamsError")
	}
	seenPeers := map[uint]bool{}
	entries := make([]*model.AddressBook, 0)
	for _, rowId := range peerIds {
		if seenPeers[rowId] {
			continue
		}
		seenPeers[rowId] = true
		peer := &model.Peer{}
		q := DB.Where("row_id = ?", rowId)
		if onlyOwned {
			q = q.Where("user_id = ?", userId)
		}
		if err := q.First(peer).Error; err != nil {
			return errors.New("ItemNotFound")
		}
		seenBooks := map[uint]bool{}
		for _, cid := range collectionIds {
			if seenBooks[cid] {
				continue
			}
			seenBooks[cid] = true
			ab := s.FromPeer(peer)
			ab.UserId = userId
			ab.CollectionId = cid
			ab.Tags = tags
			if address != nil {
				ab.Id = strings.TrimSpace(*address)
			}
			if alias != nil {
				ab.Alias = *alias
			}
			entries = append(entries, ab)
		}
	}
	return s.CreateEntries(entries)
}

// Validate a caller-supplied association, without requiring external contacts
// to exist in our device inventory.
func (s *AddressBookService) CheckPeerLink(ab *model.AddressBook, ownerOnly bool) bool {
	if ab.PeerRowId == 0 {
		return true
	}
	p := AllService.PeerService.InfoByRowId(ab.PeerRowId)
	return p.RowId > 0 && (!ownerOnly || p.UserId == ab.UserId)
}
