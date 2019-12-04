package dao

import (
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/index"
	"github.com/asdine/storm/v3/q"
	"github.com/dilshat/sms-sender/model"
	"github.com/dilshat/sms-sender/util"
	bolt "go.etcd.io/bbolt"
	"sync"
	"time"
)

type Db interface {
	Init(data interface{}) error
	One(fieldName string, value interface{}, to interface{}) error
	Update(data interface{}) error
	Save(data interface{}) error
	DeleteStruct(data interface{}) error
	Select(matchers ...q.Matcher) storm.Query
	Find(fieldName string, value interface{}, to interface{}, options ...func(q *index.Options)) error
	All(to interface{}, options ...func(*index.Options)) error
	Close() error
}

var (
	once     sync.Once
	instance Db
)

func GetClient(dbFilePath string) (Db, error) {
	var err error

	once.Do(func() {
		if !util.FileExists(dbFilePath) {
			instance, err = storm.Open(dbFilePath, storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second, ReadOnly: false}))
			if err != nil {
				return
			}
			//init db structs
			err = instance.Init(&model.Message{})
			if err != nil {
				return
			}
			err = instance.Init(&model.Recipient{})
			if err != nil {
				return
			}
		} else {
			instance, err = storm.Open(dbFilePath, storm.BoltOptions(0600, &bolt.Options{Timeout: 10 * time.Second, ReadOnly: false}))
			if err != nil {
				return
			}
		}
	})

	return instance, err
}
