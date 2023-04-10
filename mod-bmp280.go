package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/parMaster/rpid/config"
	"github.com/parMaster/rpid/storage"
	"github.com/parMaster/rpid/storage/model"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
)

const (
	HectoPascal physic.Pressure = 100 * physic.Pascal
)

type Bmp280Reporter struct {
	data         map[string][]ShortFloat
	cfg          config.BMP280
	bmp280Data   physic.Env
	bmp280Device *bmxx80.Dev
	i2cBus       i2c.BusCloser
	mx           sync.Mutex
	store        storage.Storer
}

func LoadBmp280Reporter(cfg config.BMP280, i2cBus i2c.BusCloser, store storage.Storer) (*Bmp280Reporter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Bmp280Reporter is not enabled")
	}

	data := map[string][]ShortFloat{
		"pressure": {}, // Atmospheric pressure from BMP280 in hPa
		"temp":     {}, // Temperature from BMP280 in mC
	}

	bmp280Device, err := bmxx80.NewI2C(i2cBus, cfg.Bmp280Addr, &bmxx80.DefaultOpts)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to initialize bmp280: %w", err)
	}
	b := &Bmp280Reporter{data: data, bmp280Device: bmp280Device, i2cBus: i2cBus, cfg: cfg}
	if store != nil {
		log.Printf("[DEBUG] Bmp280Reporter: using storage (%T)", store)
		b.store = store
	}
	return b, nil
}

func (r *Bmp280Reporter) Name() string {
	return "bmp280"
}

func (r *Bmp280Reporter) Collect(ctx context.Context) error {
	if err := r.bmp280Device.Sense(&r.bmp280Data); err != nil {
		return err
	}
	pressurePa := ShortFloat(r.bmp280Data.Pressure / physic.Pascal)
	tempMilliC := ShortFloat(r.bmp280Data.Temperature-physic.ZeroCelsius) / 1000000000

	r.mx.Lock()
	r.data["pressure"] = append(r.data["pressure"], pressurePa/100)
	r.data["temp"] = append(r.data["temp"], tempMilliC)
	r.mx.Unlock()

	log.Printf("[DEBUG] BMP280: %8s | %s hPa \n", r.bmp280Data.Temperature, pressurePa)

	if r.store != nil {
		err := r.store.Write(ctx, model.Data{Module: r.Name(), Topic: "pressure", Value: fmt.Sprint(pressurePa / 100)})
		if err != nil {
			return fmt.Errorf("[ERROR] Bmp280Reporter: failed to write to storage: %v", err)
		}
	}
	return nil
}

func (r *Bmp280Reporter) Report() (interface{}, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.data, nil
}
