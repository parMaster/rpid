package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {
	var err error
	conf, err := NewConfig("config.yml")
	assert.NoError(t, err)
	assert.NotNil(t, conf)
}
