package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/parMaster/htu21"
	"github.com/parMaster/rpid/config"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

type Htu21Reporter struct {
	data        historical
	cfg         config.HTU21
	htu21Data   physic.Env
	htu21Device *htu21.Dev
	i2cBus      i2c.BusCloser
	mx          sync.Mutex
}

func LoadHtu21Reporter(cfg config.HTU21, i2cBus i2c.BusCloser) (*Htu21Reporter, error) {
	data := historical{
		"humidity": {}, // Humidity from HTU21 in mRh
		"temp":     {}, // unused
	}

	htu21Device, err := htu21.NewI2C(i2cBus, cfg.Htu21Addr)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] failed to initialize htu21: %v", err)
	}
	return &Htu21Reporter{data: data, htu21Device: htu21Device, i2cBus: i2cBus, cfg: cfg}, nil
}

func (r *Htu21Reporter) Name() string {
	return "htu21"
}

func (r *Htu21Reporter) Collect() error {
	if err := r.htu21Device.Sense(&r.htu21Data); err != nil {
		return err
	}
	humidMilliRH := r.htu21Data.Humidity / 10000

	r.mx.Lock()
	r.data["humidity"] = append(r.data["humidity"], int(humidMilliRH))
	r.data["temp"] = append(r.data["temp"], int(r.htu21Data.Temperature))
	r.mx.Unlock()

	log.Printf("HTU21: %8s | %s (%d mRh) \n", r.htu21Data.Temperature, r.htu21Data.Humidity, humidMilliRH)
	return nil
}

func (r *Htu21Reporter) Report() (historical, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.data, nil
}
