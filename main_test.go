package main

import (
	"context"
	"testing"

	"github.com/parMaster/rpid/config"
	"github.com/stretchr/testify/assert"
)

func Test_SystemReporter(t *testing.T) {
	r, err := LoadSystemReporter(config.System{Enabled: true}, true)
	assert.NoError(t, err)

	ctx := context.Background()

	err = r.Collect(ctx)
	assert.NoError(t, err)

	res, err := r.Report()
	assert.NoError(t, err)

	expected := Response{
		TimeInState: map[string]int{
			"1000": 2349,
			"1100": 1911,
			"1200": 1970,
			"1300": 1799,
			"1400": 1547,
			"1500": 1143,
			"1600": 795,
			"1700": 746,
			"1800": 2726,
			"600":  200116,
			"700":  35684,
			"800":  6126,
			"900":  4051,
		},
		LoadAvg: map[string][]ShortFloat{
			"1m":  {0.12},
			"5m":  {0.24},
			"15m": {0.3},
		}}

	assert.Equal(t, expected, res)
}
