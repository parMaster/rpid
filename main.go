package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-pkgz/lgr"
	"github.com/parMaster/htu21"
	flags "github.com/umputun/go-flags"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

type historical map[string][]int

const (
	HectoPascal physic.Pressure = 100 * physic.Pascal
)

type Worker struct {
	// persistent revs counter
	revs int

	// persistent temperature measurement
	temp int

	// GPIO Fan tachymeter connected to
	// Tachymeter usually is a yellow wire in 3-pin fan connector
	fanTachPin    string
	fanControlPin string
	fanControl    gpio.PinIO
	tach          gpio.PinIn
	tempHigh      int
	tempLow       int

	// pressure and ambient temperature data from bmp280 sensor
	// to scan for i2c interfaces:
	// $ i2cdetect -l
	// i2c-4	i2c	400000002.i2c	I²C adapter
	i2cBusNumber string
	i2cBus       i2c.BusCloser
	// to find out address of the device
	// $ i2cdetect -y 4
	bmp280Addr   uint16
	bmp280Data   physic.Env
	bmp280Device *bmxx80.Dev

	// HTU21 sensor address
	htu21Addr   uint16
	htu21Data   physic.Env
	htu21Device *htu21.Dev

	// RPi file with milliCentigrades of CPU temperature
	temperatureFileName string

	// map of historical data
	data historical

	listen string

	mx sync.Mutex
}

func StartNewWorker(cfg Options, ctx context.Context) {

	data := historical{
		// Fan tachymeters
		"revs":  {}, // momentary revs/sec
		"rpm-m": {}, // rpm history by minute
		"rpm-h": {}, // rpm history by hour

		// CPU Temperature in milliCentigrades
		"temp":   {}, // momentary temp
		"temp-m": {}, // temp history by minute
		"temp-h": {}, // temp history by hour

		// Ambient temperature from BMP280
		"amb-temp-m": {}, // by minute
		"amb-temp-h": {}, // by hour

		// Atmospheric pressure from BMP280 in hPa
		"press-m": {},
		"press-h": {},

		// Relative humidity from HTU21 in mRh (0.1%)
		"rh-m": {},
		"rh-h": {},
	}

	w := &Worker{
		data:                data,
		fanTachPin:          cfg.FanTachPin,
		fanControlPin:       cfg.FanControlPin,
		tempHigh:            cfg.TempHigh,
		tempLow:             cfg.TempLow,
		listen:              cfg.Listen,
		temperatureFileName: "/sys/class/thermal/thermal_zone0/temp",
		i2cBusNumber:        cfg.I2C,
		bmp280Addr:          0x76,
		htu21Addr:           0x40,
	}

	var err error

	// Load peripheral drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	w.i2cBus, err = i2creg.Open(w.i2cBusNumber)
	if err != nil {
		log.Fatalf("[ERROR] failed to open I²C: %v", err)
	}

	w.bmp280Device, err = bmxx80.NewI2C(w.i2cBus, w.bmp280Addr, &bmxx80.DefaultOpts)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize bme280: %v", err)
	}

	w.htu21Device, err = htu21.NewI2C(w.i2cBus, w.htu21Addr)
	if err != nil {
		log.Fatalf("[ERROR] failed to initialize htu21: %v", err)
	}

	go w.controlFan(ctx)
	go w.startTach(ctx)

	go w.logEverySecond(ctx)
	go w.logEveryMinute(ctx)
	go w.logEveryHour(ctx)
	go w.startServer(ctx)

	log.Printf("Service started. Fan tach on %s, trigger on %s, listening to \"%s\"", w.fanTachPin, w.fanControlPin, w.listen)
	log.Printf("Temps cfg: low=%d˚C, high=%d˚C", w.tempLow, w.tempHigh)

	<-ctx.Done()
	time.Sleep(2 * time.Second) // wait 2 secs till tach timeout (1 sec) hits
	log.Println("[DEBUG] Closing I²C Bus on exit")
	if err := w.i2cBus.Close(); err != nil {
		log.Printf("[ERROR] Closing I²C: %e", err)
	}
}

func (w *Worker) startServer(ctx context.Context) {
	httpServer := &http.Server{
		Addr:              w.listen,
		Handler:           w.router(),
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Second,
	}

	httpServer.ListenAndServe()

	// Wait for termination signal
	<-ctx.Done()
	log.Printf("[INFO] Terminating http server")

	if err := httpServer.Close(); err != nil {
		log.Printf("[ERROR] failed to close http server, %v", err)
	}
}

func (w *Worker) router() http.Handler {
	router := chi.NewRouter()

	router.Get("/status", func(w http.ResponseWriter, r *http.Request) {

		log.Println("[DEBUG] !!! status called")

	})

	return router
}

func (w *Worker) setFanState(state bool) error {
	if err := w.fanControl.Out(gpio.Level(state)); err != nil {
		log.Printf("[ERROR] Changing fan state (%v): %e", state, err)
		return err
	}
	log.Printf("[DEBUG] Fan set to %v", gpio.Level(state))
	return nil
}

func (w *Worker) controlFan(ctx context.Context) {
	w.fanControl = gpioreg.ByName(w.fanControlPin)
	if w.fanControl == nil {
		log.Printf("[ERROR] Failed to find %s", w.fanControl)
	}

	depth := 3                    // minutes moving average
	tempHigh := w.tempHigh * 1000 // fan   activation temperature m˚C
	tempLow := w.tempLow * 1000   // fan DEactivation temperature m˚C

	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("[DEBUG] Leaving the fan ON is always safer")
			w.fanControl.Out(gpio.High)
			w.fanControl.Halt()
			return
		case <-ticker.C:
		}

		w.mx.Lock()
		lastTemp := last(w.data["temp-m"])
		tempAvg := 0
		if len(w.data["temp-m"]) >= depth {
			tempAvg = avg(w.data["temp-m"][max(0, len(w.data["temp-m"])-depth) : len(w.data["temp-m"])-1])
		}
		log.Printf("[DEBUG] temp avg: %d", tempAvg)
		w.mx.Unlock()

		// no data - keeping things safe and fan ON
		if lastTemp == 0 {
			w.setFanState(true)
			continue
		}

		// something's wrong with data - keeping things safe and fan ON
		if tempAvg == 0 {
			w.setFanState(true)
			continue
		}

		// turning fan ON - check every minute. Or when there's no temp data
		if lastTemp > tempHigh {
			w.setFanState(true)
			continue
		}

		// turning fan OFF lags 3 minutes behind
		if tempAvg < tempLow {
			w.setFanState(false)
		}
	}
}

func (w *Worker) startTach(ctx context.Context) {

	w.tach = gpioreg.ByName(w.fanTachPin)
	if w.tach == nil {
		log.Fatalf("Failed to find %s", w.fanTachPin)
	}

	// Set pin as input, with an internal pull-up resistor:
	if err := w.tach.In(gpio.PullUp, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}
	log.Printf("[DEBUG] tach %s: %s\n", w.tach, w.tach.Function())

	// Count every rev or exit
	for {
		select {
		case <-ctx.Done():
			log.Println("[DEBUG] Halting tachymeter")
			if err := w.tach.Halt(); err != nil {
				log.Printf("[ERROR] Halting tachymeter: %e", err)
			}
			return
		default:
		}
		if w.tach.WaitForEdge(time.Second) {
			w.revs++
		}
	}
}

func (w *Worker) logEverySecond(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		sysTemp, err := os.ReadFile(w.temperatureFileName)
		if err != nil {
			log.Printf("[ERROR] Can't read temperature file: %e", err)
			w.temp = 0
		} else {
			w.temp, err = strconv.Atoi(string(sysTemp[0 : len(sysTemp)-1]))
			if err != nil {
				log.Printf("[ERROR] Converting temp data: %e", err)
			}
		}

		log.Printf("[DEBUG] Temp: %d m˚C | Fan RPS/RPM: %d/%d\r\n", w.temp, w.revs, w.revs*60)

		w.mx.Lock()
		w.data["revs"] = append(w.data["revs"], w.revs*60)
		w.revs = 0
		w.data["temp"] = append(w.data["temp"], w.temp)
		w.mx.Unlock()
	}
}

// Aggregate measurements by second to data by minute
func (w *Worker) logEveryMinute(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		if err := w.bmp280Device.Sense(&w.bmp280Data); err != nil {
			log.Fatal(err)
		}
		pressureHPa := w.bmp280Data.Pressure / HectoPascal
		tempMilliC := w.bmp280Data.Temperature / 10000000

		if err := w.htu21Device.Sense(&w.htu21Data); err != nil {
			log.Fatal(err)
		}
		humidMilliRH := w.htu21Data.Humidity / 10000

		w.mx.Lock()
		w.data["rpm-m"] = append(w.data["rpm-m"], avg(w.data["revs"]))
		w.data["revs"] = []int{}

		w.data["temp-m"] = append(w.data["temp-m"], avg(w.data["temp"]))
		w.data["temp"] = []int{}

		w.data["amb-temp-m"] = append(w.data["amb-temp-m"], int(tempMilliC))
		w.data["press-m"] = append(w.data["press-m"], int(pressureHPa))
		w.data["rh-m"] = append(w.data["rh-m"], int(humidMilliRH))

		log.Printf("CPU: %d m˚C\r\n", last(w.data["temp-m"]))
		log.Printf("Fan: %d rpm\r\n", last(w.data["rpm-m"]))
		log.Printf("BMP280: %8s | %d hPa \n", w.bmp280Data.Temperature, pressureHPa)
		log.Printf("HTU21: %8s | %s (%d mRh) \n", w.htu21Data.Temperature, w.htu21Data.Humidity, humidMilliRH)
		w.mx.Unlock()
	}
}

func (w *Worker) aggregateHourly(source, dest string) {
	w.data[dest] = append(w.data[dest], avg(w.data[source][max(0, len(w.data[source])-60):len(w.data[source])-1]))
}

// Aggregate measurements by second to data hourly
func (w *Worker) logEveryHour(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		w.mx.Lock()
		w.aggregateHourly("rpm-m", "rpm-h")
		w.aggregateHourly("temp-m", "temp-h")
		w.aggregateHourly("amb-temp-m", "amb-temp-h")
		w.aggregateHourly("press-m", "press-h")
		w.aggregateHourly("rh-m", "rh-h")

		log.Print("*** Hourly \r\n")
		log.Printf("CPU: %d m˚C\r\n", last(w.data["temp-h"]))
		log.Printf("Fan: %d rpm\r\n", last(w.data["rpm-h"]))
		log.Printf("Ambient Temp: %d m˚C\r\n", last(w.data["amb-temp-h"]))
		log.Printf("Atmospheric pressure: %d hPa\r\n", last(w.data["press-h"]))
		log.Printf("Relative Humidity: %d mRh\r\n", last(w.data["rh-h"]))
		w.mx.Unlock()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func last(slice []int) int {
	if len(slice) == 0 {
		return 0
	}
	return slice[len(slice)-1]
}

func avg(slice []int) int {
	if len(slice) == 0 {
		return 0
	}
	sum := 0
	for _, v := range slice {
		sum += v
	}
	return int(sum / len(slice))
}

type Options struct {
	Listen        string `long:"listen" env:"LISTEN" default:"pi4.local:8095" description:"Port for http server to listen to"`
	FanTachPin    string `long:"tach-pin" env:"TACH" default:"GPIO15" description:"GPIO with fan tachymeter connected"`
	FanControlPin string `long:"control-pin" env:"CONTROL" default:"GPIO18" description:"GPIO with fan control connected - base of the key transistor"`
	TempHigh      int    `long:"temp-high" env:"TEMPHIGH" default:"45" description:"Fan activation temperature"`
	TempLow       int    `long:"temp-low" env:"TEMPLOW" default:"40" description:"Fan deactivation temperature"`
	I2C           string `long:"i2cbus" env:"I2C" default:"4" description:"I2C bus number"`
	Dbg           bool   `long:"dbg" env:"DEBUG" description:"show debug info"`
}

var opts Options

func main() {
	// Parsing cmd parameters
	p := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		p.WriteHelp(os.Stderr)
		os.Exit(2)
	}

	// Logger setup
	logOpts := []lgr.Option{
		lgr.LevelBraces,
		lgr.StackTraceOnError,
	}
	if opts.Dbg {
		logOpts = append(logOpts, lgr.Debug)
	}
	lgr.SetupStdLogger(logOpts...)

	// Graceful termination
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic:\n%v", x)
			panic(x)
		}

		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("Shutdown signal received\n*********************************")
		cancel()
	}()

	StartNewWorker(opts, ctx)
}
