package db

import "github.com/jinzhu/gorm"

type Storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{
		db: db,
	}
}
