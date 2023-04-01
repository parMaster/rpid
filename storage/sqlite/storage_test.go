package sqlite

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/parMaster/rpid/storage/model"
	"github.com/stretchr/testify/assert"
)

func Test_SqliteStorage(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	store, err := NewStorage(ctx, "test.db", false)
	if err != nil {
		log.Printf("[ERROR] Failed to open SQLite storage: %e", err)
	}

	testRecord := model.Data{
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
	err = store.Write(ctx, model.Data{Module: "testModule", Topic: "", Value: "testValue"})
	assert.Error(t, err)

	// empty value is allowed
	err = store.Write(ctx, model.Data{Module: "testModule", Topic: "testTopic", Value: ""})
	assert.NoError(t, err)

	// Test if the date time is set to the current time if it is not set.
	dt := time.Now().Format("2006-01-02 15:04")
	err = store.Write(ctx, model.Data{Module: "testModule", Topic: "testTopic", Value: "testValue"})
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

func Test_SqliteStorage_View(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	store, err := NewStorage(ctx, "view.db", false)
	if err != nil {
		log.Printf("[ERROR] Failed to open SQLite storage: %e", err)
	}

	records := []model.Data{
		{Module: "view", DateTime: "2022-03-30 00:00", Topic: "temp", Value: "36000"},
		{Module: "view", DateTime: "2022-03-30 00:01", Topic: "temp", Value: "36100"},
		{Module: "view", DateTime: "2022-03-30 00:02", Topic: "temp", Value: "36200"},
		{Module: "view", DateTime: "2022-03-30 00:00", Topic: "rpm", Value: "100"},
		{Module: "view", DateTime: "2022-03-30 00:01", Topic: "rpm", Value: "200"},
		{Module: "view", DateTime: "2022-03-30 00:02", Topic: "rpm", Value: "300"},
	}

	for _, r := range records {
		err = store.Write(ctx, r)
		assert.NoError(t, err)
	}

	// test if the view is created

	viewExpected := map[string]map[string]string{
		"temp": {
			"2022-03-30 00:00": "36000",
			"2022-03-30 00:01": "36100",
			"2022-03-30 00:02": "36200",
		},
		"rpm": {
			"2022-03-30 00:00": "100",
			"2022-03-30 00:01": "200",
			"2022-03-30 00:02": "300",
		},
	}

	view, err := store.View(ctx, "view") // create the view
	assert.NoError(t, err)
	assert.Equal(t, viewExpected, view)

}

func Test_SqliteStorage_readOnly(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	store, err := NewStorage(ctx, "test.db", true)
	if err != nil {
		log.Printf("[ERROR] Failed to open SQLite storage: %e", err)
	}

	err = store.Write(ctx, model.Data{Module: "testModule", Topic: "testTopic", Value: "testValue"})
	assert.Error(t, err)
	assert.Equal(t, "storage is read-only", err.Error())

	_, err = NewStorage(ctx, "test_notcreated.db", true)
	assert.Error(t, err)
	assert.Equal(t, "storage does not exist and is read-only", err.Error())

}
