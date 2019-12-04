package dao

import (
	"github.com/dilshat/sms-sender/model"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

const (
	SENDER  = "Awesome"
	TEXT    = "Hello World!"
	SENDER2 = "Sky"
	TEXT2   = "Hello Earth!"
)

var (
	ID1 uint32
	ID2 uint32
)

func prepareDB(t errorHandler) (Db, func()) {
	db, cleanup := createDB(t)

	//populate db
	msg := &model.Message{Sender: SENDER, Text: TEXT, CreatedAt: time.Now()}
	err := db.Save(msg)
	if err != nil {
		log.Fatal(err)
	}
	ID1 = msg.Id
	msg = &model.Message{Sender: SENDER2, Text: TEXT2, CreatedAt: time.Now().Add(-25 * time.Hour)}
	err = db.Save(msg)
	if err != nil {
		log.Fatal(err)
	}
	ID2 = msg.Id

	return db, cleanup
}

type errorHandler interface {
	Error(args ...interface{})
}

func TestMessageDao_Create(t *testing.T) {
	db, cleanup := createDB(t)
	defer cleanup()
	msgDao := NewMessageDao(db)

	id, err := msgDao.Create(TEXT, SENDER)

	require.NoError(t, err)
	require.True(t, id > 0)
}

func TestMessageDao_GetOneById(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()
	msgDao := NewMessageDao(db)

	msg, err := msgDao.GetOneById(ID1)

	require.NoError(t, err)
	require.NotEmpty(t, msg)
	require.Equal(t, ID1, msg.Id)
}

func TestMessageDao_GetAll(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()
	msgDao := NewMessageDao(db)

	all, err := msgDao.GetAll()

	require.NoError(t, err)
	require.Equal(t, 2, len(all))
}

func TestMessageDao_RemoveOlderThanDays(t *testing.T) {
	db, cleanup := prepareDB(t)
	defer cleanup()
	msgDao := NewMessageDao(db)

	err := msgDao.RemoveOlderThanDays(1)

	require.NoError(t, err)

	all, _ := msgDao.GetAll()
	require.Equal(t, 1, len(all))
}
