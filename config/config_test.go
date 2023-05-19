package config

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadConfig(t *testing.T) {

	// Expected default config:
	expected := Parameters{
		Server: Server{
			Listen: ":8095",
			Dbg:    false,
		},
		Fan: Fan{
			TachPin:    "GPIO15",
			ControlPin: "GPIO18",
			High:       48,
			Low:        40,
		},
		Storage: Storage{
			Type: "sqlite",
			Path: "file:/etc/rpid/data.db?mode=rwc&_journal_mode=WAL",
		},
		Modules: Modules{
			BMP280: BMP280{
				Enabled:    true,
				Bmp280Addr: 118,
			},
			HTU21: HTU21{
				Enabled:   true,
				Htu21Addr: 64,
			},
			System: System{
				Enabled: true,
			},
			I2C: "4",
		},
	}

	var conf *Parameters
	var err error
	conf, err = NewConfig("config.yml")
	if err != nil {
		log.Fatalf("[ERROR] can't load config, %s", err)
	}
	assert.Equal(t, expected, *conf)
}
