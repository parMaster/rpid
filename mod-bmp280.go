package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/parMaster/rpid/config"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
)

type Bmp280Reporter struct {
	data         historical
	cfg          config.BMP280
	bmp280Data   physic.Env
	bmp280Device *bmxx80.Dev
	i2cBus       i2c.BusCloser
	mx           sync.Mutex
}

func LoadBmp280Reporter(cfg config.BMP280, i2cBus i2c.BusCloser) (*Bmp280Reporter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Bmp280Reporter is not enabled")
	}

	data := historical{
		"pressure": {}, // Atmospheric pressure from BMP280 in hPa
		"temp":     {}, // Temperature from BMP280 in mC
	}

	bmp280Device, err := bmxx80.NewI2C(i2cBus, cfg.Bmp280Addr, &bmxx80.DefaultOpts)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to initialize bmp280: %w", err)
	}
	return &Bmp280Reporter{data: data, bmp280Device: bmp280Device, i2cBus: i2cBus, cfg: cfg}, nil
}

func (r *Bmp280Reporter) Name() string {
	return "bmp280"
}

func (r *Bmp280Reporter) Collect() error {
	if err := r.bmp280Device.Sense(&r.bmp280Data); err != nil {
		return err
	}
	pressureHPa := r.bmp280Data.Pressure / HectoPascal
	tempMilliC := int64(r.bmp280Data.Temperature-physic.ZeroCelsius) / 1000000

	r.mx.Lock()
	r.data["pressure"] = append(r.data["pressure"], int(pressureHPa))
	r.data["temp"] = append(r.data["temp"], int(tempMilliC))
	r.mx.Unlock()

	log.Printf("BMP280: %8s | %d hPa \n", r.bmp280Data.Temperature, pressureHPa)
	return nil
}

func (r *Bmp280Reporter) Report() (interface{}, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.data, nil
}
