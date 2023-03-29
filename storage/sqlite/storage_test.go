package sqlite

import (
	"context"
	"log"
	"testing"
	"time"

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
		Module:   "testModule",
		DateTime: "2019-01-01 00:00",
		Topic:    "testTopic",
		Value:    "testValue",
	}

	store.Cleanup(testRecord.Module)

	// write a record
	err = store.Write(ctx, testRecord)
	assert.Nil(t, err)

	// read the record
	data, err := store.Read(ctx, testRecord.Module)
	assert.Equal(t, 1, len(data))
	assert.Nil(t, err)
	assert.Equal(t, data[0], testRecord)

	// write another record
	err = store.Write(ctx, testRecord)
	assert.Nil(t, err)
	data, err = store.Read(ctx, testRecord.Module)
	assert.Equal(t, 2, len(data))
	assert.Nil(t, err)
	assert.Equal(t, data[1], testRecord)

	// test if the module is not active (no such table)
	data, err = store.Read(ctx, "notable")
	assert.Nil(t, data)
	assert.Error(t, err)

	// empty topic is not allowed
	err = store.Write(ctx, storage.Data{Module: "testModule", Topic: "", Value: "testValue"})
	assert.Error(t, err)

	// empty value is allowed
	err = store.Write(ctx, storage.Data{Module: "testModule", Topic: "testTopic", Value: ""})
	assert.NoError(t, err)

	// Test if the date time is set to the current time if it is not set.
	dt := time.Now().Format("2006-01-02 15:04")
	err = store.Write(ctx, storage.Data{Module: "testModule", Topic: "testTopic", Value: "testValue"})
	assert.NoError(t, err)
	savedValues, err := store.Read(ctx, "testModule")
	assert.NoError(t, err)
	assert.Equal(t, dt, savedValues[len(savedValues)-1].DateTime)

	v, _ := store.Read(ctx, "testModule")
	n := 100
	// n records
	for i := 0; i < n; i++ {
		err = store.Write(ctx, testRecord)
		assert.NoError(t, err)
	}
	vn, _ := store.Read(ctx, "testModule")
	assert.Equal(t, n, len(vn)-len(v))
}
