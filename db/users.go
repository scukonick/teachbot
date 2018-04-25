package db

import (
	"time"

	"github.com/scukonick/teachbot/db/structs"
	"github.com/jinzhu/gorm"
)

func (s *Storage) GetUserByTgID(tgID int64) (*structs.User, error) {
	u := &structs.User{}
	err := s.db.Where("tg_id = ?", tgID).First(u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Storage) UpsertUserByTgID(u *structs.User) (*structs.User, error) {
	existing := &struct {
		ID        int32
		CreatedAt time.Time
	}{}

	err := s.db.Raw("SELECT id, created_at FROM users WHERE tg_id = ?", u.TgID).
		Scan(existing).Error

	if err == gorm.ErrRecordNotFound {
		err = s.db.Create(u).Error
		if err != nil {
			return nil, err
		}
		return u, nil

	} else if err != nil {
		return nil, err
	}

	u.ID = existing.ID
	u.CreatedAt = existing.CreatedAt
	err = s.db.Save(u).Error
	if err != nil {
		return nil, err
	}

	return u, err
}
