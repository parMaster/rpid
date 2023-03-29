package sqlite

import (
	"context"
	"log"
	"testing"

	"github.com/parMaster/rpid/storage"
	"github.com/stretchr/testify/assert"
)

func Test_SqliteStorage(t *testing.T) {

	var err error
	store, err := NewStorage("/etc/rpid/data.db")
	if err != nil {
		log.Printf("[ERROR] Failed to open SQLite storage: %e", err)
	}

	testRecord := storage.Data{
		Module:   "testqwe",
		DateTime: "2019-01-01 00:00:00",
		Topic:    "testTopic",
		Value:    "testValue",
	}

	ctx := context.Background()

	store.Cleanup(testRecord.Module)

	err = store.Write(ctx, testRecord)
	assert.Nil(t, err)

	data, err := store.Read(ctx, testRecord.Module)
	assert.Equal(t, 1, len(data))
	assert.Nil(t, err)
	assert.Equal(t, data[0], testRecord)

	err = store.Write(ctx, testRecord)
	assert.Nil(t, err)
	data, err = store.Read(ctx, testRecord.Module)
	assert.Equal(t, 2, len(data))
	assert.Nil(t, err)
	assert.Equal(t, data[1], testRecord)

	store.Close()
}
