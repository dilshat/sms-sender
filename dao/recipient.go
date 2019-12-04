package dao

import (
	"github.com/dilshat/sms-sender/model"
	"github.com/asdine/storm/v3/q"
	"time"
)

type RecipientDao interface {
	//Create creates recipient record and returns its id
	Create(messageId uint32, phone string) (uint32, error)
	//UpdateSubmitStatus updates status and delivery id of recipient record with the given id
	UpdateSubmitStatus(id uint32, deliverId uint64, status string) error
	//UpdateDeliverStatus updates status of recipient record with the delivery id
	UpdateDeliverStatus(deliverId uint64, status string) (uint32, string, error)
	//GetOneByMessageIdAndPhone returns a recipient with the given message id and phone
	GetOneByMessageIdAndPhone(messageId uint32, phone string) (model.Recipient, error)
	//GetAllByMessageId returns all recipients with the given message id
	GetAllByMessageId(messageId uint32) ([]model.Recipient, error)
	//GetAll returns all recipients
	GetAll() ([]model.Recipient, error)
	//RemoveOlderThanDays removes all recipients older that {days}
	RemoveOlderThanDays(days int) error
}

func NewRecipientDao(db Db) RecipientDao {
	return &recipientDao{db: db}
}

type recipientDao struct {
	db Db
}

func (r recipientDao) RemoveOlderThanDays(days int) error {
	err := r.db.Select(q.Lt("CreatedAt", time.Now().Add(-24*time.Duration(days)*time.Hour))).Delete(&model.Recipient{})
	if err != nil && err.Error() != "not found" {
		return err
	}
	return nil
}

func (r recipientDao) Create(messageId uint32, phone string) (uint32, error) {
	recipient := &model.Recipient{MessageId: messageId, Phone: phone, Status: model.NEW, CreatedAt: time.Now()}
	err := r.db.Save(recipient)
	return recipient.Id, err
}

func (r recipientDao) UpdateSubmitStatus(id uint32, deliverId uint64, status string) error {
	//update status based on SUBMIT_SM_RESP status
	var recipient model.Recipient
	err := r.db.One("Id", id, &recipient)
	if err != nil {
		return err
	}
	recipient.DeliverId = deliverId
	recipient.Status = status
	return r.db.Update(&recipient)
}

func (r recipientDao) UpdateDeliverStatus(deliverId uint64, status string) (uint32, string, error) {
	//update status based on DELIVER_SM
	var recipient model.Recipient
	err := r.db.One("DeliverId", deliverId, &recipient)
	if err != nil {
		return 0, "", err
	}
	recipient.Status = status
	err = r.db.Update(&recipient)
	return recipient.MessageId, recipient.Phone, err
}

func (r recipientDao) GetOneByMessageIdAndPhone(messageId uint32, phone string) (model.Recipient, error) {
	var matchers []q.Matcher
	matchers = append(matchers, q.Eq("MessageId", messageId))
	matchers = append(matchers, q.Eq("Phone", phone))
	var recipients []model.Recipient
	err := r.db.Select(matchers...).Limit(1).Find(&recipients)
	var recipient model.Recipient
	if err != nil {
		return recipient, err
	}
	if len(recipients) > 0 {
		recipient = recipients[0]
	}

	return recipient, err
}

func (r recipientDao) GetAllByMessageId(messageId uint32) (recipients []model.Recipient, err error) {
	err = r.db.Find("MessageId", messageId, &recipients)
	return
}

func (r recipientDao) GetAll() (recipients []model.Recipient, err error) {
	err = r.db.All(&recipients)
	return
}
