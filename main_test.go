package main

import (
	"context"
	"log"
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

func Test_LoadConfig(t *testing.T) {

	// Expected default config:
	expected := config.Parameters{
		Server: config.Server{
			Listen: ":8095",
			Dbg:    false,
		},
		Fan: config.Fan{
			TachPin:    "GPIO15",
			ControlPin: "GPIO18",
			High:       45,
			Low:        40,
		},
		Storage: config.Storage{
			Type: "sqlite",
			Path: "/etc/rpid/data.db",
		},
		Modules: config.Modules{
			BMP280: config.BMP280{
				Enabled:    true,
				Bmp280Addr: 118,
			},
			HTU21: config.HTU21{
				Enabled:   true,
				Htu21Addr: 64,
			},
			System: config.System{
				Enabled: true,
			},
			I2C: "4",
		},
	}

	var conf *config.Parameters
	var err error
	conf, err = config.NewConfig("config/config.yml")
	if err != nil {
		log.Fatalf("[ERROR] can't load config, %s", err)
	}
	assert.Equal(t, expected, *conf)
}
