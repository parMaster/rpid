package sqlite

import (
	"context"
	"log"
	"testing"

	"github.com/parMaster/rpid/storage"
	"github.com/stretchr/testify/assert"
)

func Test_SqliteStorage(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	store, err := NewStorage(ctx, "test.db")
	if err != nil {
		log.Printf("[ERROR] Failed to open SQLite storage: %e", err)
	}

	testRecord := storage.Data{
		Module:   "testqwe",
		DateTime: "2019-01-01 00:00:00",
		Topic:    "testTopic",
		Value:    "testValue",
	}

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

	notableRecord := storage.Data{
		Module:   "notable",
		DateTime: "2019-01-01 00:00:00",
		Topic:    "testTopic",
		Value:    "testValue",
	}
	data, err = store.Read(ctx, notableRecord.Module)
	assert.Nil(t, data)
	assert.Error(t, err)
}
