package db

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/scukonick/teachbot/db/structs"
)

func (s *Storage) GetImageTaskForUser(userID int32) (*structs.Task, error) {
	q := `
SELECT * FROM tasks WHERE id NOT IN
(SELECT task_id FROM user_tasks WHERE user_id = ?)
AND image IS NOT NULL AND image != ''
ORDER BY RANDOM() LIMIT 1`
	resp := &structs.Task{}
	err := s.db.Raw(q, userID).Scan(resp).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "db lookup failed")
	}

	return resp, nil
}

func (s *Storage) GetTextTaskForUser(userID int32) (*structs.Task, error) {
	q := `
SELECT * FROM tasks WHERE id NOT IN
(SELECT task_id FROM user_tasks WHERE user_id = ?)
AND (image IS NULL OR image = '')
ORDER BY RANDOM() LIMIT 1`
	resp := &structs.Task{}
	err := s.db.Raw(q, userID).Scan(resp).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, errors.Wrap(err, "db lookup failed")
	}

	return resp, nil
}

func (s *Storage) CreateUserTask(userID, taskID int32) error {
	t := &structs.UserTask{
		UserID: userID,
		TaskID: taskID,
	}

	err := s.db.Create(t).Error
	if err != nil {
		return errors.Wrap(err, "db insert failed")
	}

	return nil
}
