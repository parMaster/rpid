package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/parMaster/htu21"
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
	tachPin       string
	fanTriggerPin string
	fanTrigger    gpio.PinIO
	tach          gpio.PinIn

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

	mx sync.Mutex
}

func StartNewWorker() *Worker {

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
		tachPin:             "GPIO15",
		fanTriggerPin:       "GPIO18",
		temperatureFileName: "/sys/class/thermal/thermal_zone0/temp",
		i2cBusNumber:        "4",
		bmp280Addr:          0x76,
		htu21Addr:           0x40,
	}

	go w.controlFan()
	go w.startTach()

	go w.logEverySecond()
	go w.logEveryMinute()
	go w.logEveryHour()

	var err error
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

	log.Printf("Service started. Fan tach on %s, trigger on %s", w.tachPin, w.fanTriggerPin)

	return w
}

func (w *Worker) Halt() {

	log.Println("[DEBUG] Halting tachymeter")
	err := w.tach.Halt()
	if err != nil {
		log.Printf("[ERROR] Halting tachymeter: %e", err)
	}

	log.Println("[DEBUG] Closing I²C Bus")
	w.i2cBus.Close()
	if err != nil {
		log.Printf("[ERROR] Closing I²C: %e", err)
	}

	log.Println("[DEBUG] Leaving the fan ON is always safer")
	w.fanTrigger.Out(gpio.High)
	w.fanTrigger.Halt()
}

func (w *Worker) controlFan() {
	w.fanTrigger = gpioreg.ByName(w.fanTriggerPin)
	if w.fanTrigger == nil {
		log.Printf("[ERROR] Failed to find %s", w.fanTrigger)
	}

	// Setting fan High for now
	if err := w.fanTrigger.Out(gpio.High); err != nil {
		log.Printf("[ERROR] turning fan ON: %e", err)
	}
}

func (w *Worker) startTach() {
	// Lookup a pin by its number
	w.tach = gpioreg.ByName(w.tachPin)
	if w.tach == nil {
		log.Fatalf("Failed to find %s", w.tachPin)
	}

	// Set it as input, with an internal pull-up resistor:
	if err := w.tach.In(gpio.PullUp, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	log.Printf("[DEBUG] tach %s: %s\n", w.tach, w.tach.Function())

	for {
		w.tach.WaitForEdge(-1)
		w.revs++
	}
}

func (w *Worker) logEverySecond() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		log.Printf("[DEBUG] Fan RPS/RPM: %d/%d \r\n", w.revs, w.revs*60)

		w.mx.Lock()
		w.data["revs"] = append(w.data["revs"], w.revs*60)
		w.revs = 0

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

		w.data["temp"] = append(w.data["temp"], w.temp)
		w.mx.Unlock()

		log.Printf("[DEBUG] Temperature (milli˚C): %d\r\n", w.temp)
	}
}

// Aggregate measurements by second to data by minute
func (w *Worker) logEveryMinute() {
	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {

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
		w.mx.Unlock()

		log.Printf("CPU: %d m˚C\r\n", last(w.data["temp-m"]))
		log.Printf("Fan: %d rpm\r\n", last(w.data["rpm-m"]))
		log.Printf("BMP280: %8s | %d hPa \n", w.bmp280Data.Temperature, pressureHPa)
		log.Printf("HTU21: %8s | %s (%d mRh) \n", w.htu21Data.Temperature, w.htu21Data.Humidity, humidMilliRH)
	}
}

func (w *Worker) aggregateHourly(source, dest string) {
	w.data[dest] = append(w.data[dest], avg(w.data[source][max(0, len(w.data[source])-60):len(w.data[source])-1]))
}

// Aggregate measurements by second to data hourly
func (w *Worker) logEveryHour() {
	ticker := time.NewTicker(60 * time.Minute)
	for range ticker.C {
		w.mx.Lock()
		w.aggregateHourly("rpm-m", "rpm-h")
		w.aggregateHourly("temp-m", "temp-h")
		w.aggregateHourly("amb-temp-m", "amb-temp-h")
		w.aggregateHourly("press-m", "press-h")
		w.aggregateHourly("rh-m", "rh-h")
		w.mx.Unlock()

		log.Print("*** Hourly \r\n")
		log.Printf("CPU: %d m˚C\r\n", last(w.data["temp-h"]))
		log.Printf("Fan: %d rpm\r\n", last(w.data["rpm-h"]))
		log.Printf("Ambient Temp: %d m˚C\r\n", last(w.data["amb-temp-h"]))
		log.Printf("Atmospheric pressure: %d hPa\r\n", last(w.data["press-h"]))
		log.Printf("Relative Humidity: %d mRh\r\n", last(w.data["rh-h"]))
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

func main() {
	f, err := os.OpenFile("/var/log/rpid.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error opening file: %e", err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)

	logOpts := []lgr.Option{
		// lgr.Debug,
		lgr.LevelBraces,
		lgr.StackTraceOnError,
		lgr.Out(mw),
	}
	lgr.SetupStdLogger(logOpts...)

	// Load peripheral drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	w := StartNewWorker()

	<-termChan

	log.Println("Shutdown signal received\n*********************************")
	w.Halt()
}
