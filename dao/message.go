package dao

import (
	"github.com/asdine/storm/q"
	"github.com/dilshat/sms-sender/model"
	"time"
)

type MessageDao interface {
	//Create creates message record and returns its id
	Create(text, sender string) (uint32, error)
	//GetOneById returns message by id
	GetOneById(id uint32) (model.Message, error)
	//GetAll returns all messages
	GetAll() ([]model.Message, error)
	//RemoveOlderThanDays removes all messages older than {days}
	RemoveOlderThanDays(days int) error
}

func NewMessageDao(db Db) MessageDao {
	return &messageDao{db: db}
}

type messageDao struct {
	db Db
}

func (d messageDao) RemoveOlderThanDays(days int) error {
	err := d.db.Select(q.Lt("CreatedAt", time.Now().Add(-24*time.Duration(days)*time.Hour))).Delete(&model.Message{})
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
}

func (d messageDao) GetOneById(id uint32) (recipient model.Message, err error) {
	err = d.db.One("Id", id, &recipient)
	return
}

func (d messageDao) GetAll() (messages []model.Message, err error) {
	err = d.db.All(&messages)
	return
}

func (d messageDao) Create(text, sender string) (uint32, error) {
	msg := &model.Message{Sender: sender, Text: text, CreatedAt: time.Now()}
	err := d.db.Save(msg)
	return msg.Id, err
}
