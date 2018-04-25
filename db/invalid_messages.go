package db

import "github.com/scukonick/teachbot/db/structs"

func (s *Storage) CreateInvalidMessage(userID int32, message string) error {
	msg := &structs.InvalidMessage{
		UserID:  userID,
		Message: message,
	}

	err := s.db.Create(msg).Error

	return err
}
