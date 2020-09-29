package dao

import (
	"testing"
	"time"

	"github.com/dilshat/sms-sender/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	MSG_ID1    = uint32(123)
	MSG_ID2    = uint32(321)
	PHONE1     = "996YYYAABBCC"
	PHONE2     = "999ZZZXXXXXX"
	DELIVER_ID = "1234"
)

func prepareDB2(t errorHandler) (Db, func()) {
	db, cleanup := createDB(t)

	//populate db
	msg := &model.Recipient{MessageId: MSG_ID1, Phone: PHONE1, Id: ID1, CreatedAt: time.Now()}
	err := db.Save(msg)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	ID1 = msg.Id
	msg = &model.Recipient{MessageId: MSG_ID2, Phone: PHONE2, Id: ID2, CreatedAt: time.Now().Add(-25 * time.Hour)}
	err = db.Save(msg)
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	ID2 = msg.Id

	return db, cleanup
}

func TestRecipientDao_Create(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()
	recDao := NewRecipientDao(db)

	id, err := recDao.Create(MSG_ID1, PHONE1)

	require.NoError(t, err)
	require.True(t, id > 0)
}

func TestRecipientDao_GetAll(t *testing.T) {
	db, cleanup := prepareDB2(t)
	defer cleanup()
	recDao := NewRecipientDao(db)

	all, err := recDao.GetAll()

	require.NoError(t, err)
	require.Equal(t, 2, len(all))
}

func TestRecipientDao_GetAllByMessageId(t *testing.T) {
	db, cleanup := prepareDB2(t)
	defer cleanup()
	recDao := NewRecipientDao(db)

	all, err := recDao.GetAllByMessageId(MSG_ID2)

	require.NoError(t, err)
	require.NotEmpty(t, all)
	require.Equal(t, PHONE2, all[0].Phone)
}

func TestRecipientDao_GetOneByMessageIdAndPhone(t *testing.T) {
	db, cleanup := prepareDB2(t)
	defer cleanup()
	recDao := NewRecipientDao(db)

	one, err := recDao.GetOneByMessageIdAndPhone(MSG_ID1, PHONE1)

	require.NoError(t, err)
	require.NotEmpty(t, one)
	require.Equal(t, ID1, one.Id)
}

func TestRecipientDao_UpdateSubmitStatus(t *testing.T) {
	db, cleanup := prepareDB2(t)
	defer cleanup()
	recDao := NewRecipientDao(db)

	err := recDao.UpdateSubmitStatus(ID1, DELIVER_ID, model.ACCEPTD)

	require.NoError(t, err)

	one, _ := recDao.GetOneByMessageIdAndPhone(MSG_ID1, PHONE1)

	require.Equal(t, DELIVER_ID, one.DeliverId)
	require.Equal(t, model.ACCEPTD, one.Status)
}

func TestRecipientDao_UpdateDeliverStatus(t *testing.T) {
	db, cleanup := prepareDB2(t)
	defer cleanup()
	recDao := NewRecipientDao(db)
	_ = recDao.UpdateSubmitStatus(ID1, DELIVER_ID, model.ACCEPTD)

	msgId, phone, err := recDao.UpdateDeliverStatus(DELIVER_ID, model.DELIVRD)

	require.True(t, len(phone) > 0)
	require.NoError(t, err)
	require.True(t, msgId > 0)

	one, _ := recDao.GetOneByMessageIdAndPhone(MSG_ID1, PHONE1)

	require.Equal(t, DELIVER_ID, one.DeliverId)
	require.Equal(t, model.DELIVRD, one.Status)
}

func TestRecipientDao_RemoveOlderThanDays(t *testing.T) {
	db, cleanup := prepareDB2(t)
	defer cleanup()
	recDao := NewRecipientDao(db)

	err := recDao.RemoveOlderThanDays(1)

	require.NoError(t, err)

	all, _ := recDao.GetAll()

	require.True(t, len(all) == 1)
	require.Equal(t, MSG_ID1, all[0].MessageId)
}
