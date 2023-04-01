package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Parameters is the main configuration struct
type Parameters struct {
	Server  Server  `yaml:"server"`
	Fan     Fan     `yaml:"fan"`
	Modules Modules `yaml:"modules"`
	Storage Storage `yaml:"storage"`
}

type Modules struct {
	BMP280 BMP280 `yaml:"bmp280"`
	HTU21  HTU21  `yaml:"htu21"`
	System System `yaml:"system"`
	// to scan for i2c interfaces:
	// $ i2cdetect -l
	// i2c-4	i2c	400000002.i2c	I²C adapter
	I2C string `yaml:"i2c"`
}

type Fan struct {
	// GPIO Fan tachymeter connected to
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	TachPin string `yaml:"tachPin"`
	// GPIO Fan control connected to (base of transistor)
	ControlPin string `yaml:"controlPin"`
	High       int    `yaml:"high"` // Fan activation temperature ˚C
	Low        int    `yaml:"low"`  // Fan deactivation temperature ˚C
}

type Server struct {
	Listen string `yaml:"listen"` // Address or/and Port for http server to listen to
	Dbg    bool   `yaml:"-"`
}

type Storage struct {
	// Type of storage to use
	// Currently supported: sqlite, memory
	Type string `yaml:"type"`
	// Path to the database file
	// Used only with sqlite storage
	Path string `yaml:"path"`
	// ReadOnly mode - no writes to the database, no tables creation
	ReadOnly bool `yaml:"readOnly"`
}

// to find out address of the device, use i2cdetect with -y option with the bus number
// $ i2cdetect -y 4
type BMP280 struct {
	Enabled    bool   `yaml:"enabled,omitempty"`
	Bmp280Addr uint16 `yaml:"addr,omitempty"`
}
type HTU21 struct {
	Enabled   bool   `yaml:"enabled,omitempty"`
	Htu21Addr uint16 `yaml:"addr,omitempty"`
}
type System struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

// New creates a new Parameters from the given file
func NewConfig(fname string) (*Parameters, error) {
	p := &Parameters{}
	data, err := os.ReadFile(fname)
	if err != nil {
		log.Printf("[ERROR] can't read config %s: %e", fname, err)
		return nil, fmt.Errorf("can't read config %s: %w", fname, err)
	}
	if err = yaml.Unmarshal(data, &p); err != nil {
		log.Printf("[ERROR] failed to parse config %s: %e", fname, err)
		return nil, fmt.Errorf("failed to parse config %s: %w", fname, err)
	}
	log.Printf("[DEBUG] config: %+v", p)
	return p, nil
}
