package dao

import (
	"github.com/asdine/storm/v3"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGetClient(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	require.NoError(t, err)

	dbPath := filepath.Join(dir, "storm.db")
	db, err := GetClient(dbPath)
	require.NoError(t, err)

	defer func() {
		db.Close()
		os.RemoveAll(dbPath)
	}()

	require.FileExists(t, dbPath, "Expected that db file exists")
}

func TestGetClientSingleton(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	require.NoError(t, err)

	dbPath := filepath.Join(dir, "storm.db")
	db, err := GetClient(dbPath)
	require.NoError(t, err)

	defer func() {
		db.Close()
		os.RemoveAll(dbPath)
	}()

	db2, err := GetClient(dbPath)
	require.NoError(t, err)

	require.Equal(t, db, db2)
}

func createDB(t errorHandler) (Db, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "storm")
	if err != nil {
		t.Error(err)
	}
	db, err := storm.Open(filepath.Join(dir, "storm.db"))
	if err != nil {
		t.Error(err)
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestGetClientExistingDb(t *testing.T) {
	db, _ := createDB(t)
	stormDb := db.(*storm.DB)
	dbPath := stormDb.Bolt.Path()
	_=stormDb.Close()
	defer os.RemoveAll(dbPath)

	clnt, err := GetClient(dbPath)

	require.NoError(t, err)
	require.NotEmpty(t,clnt)

}
