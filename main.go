package main

import (
	"context"
	_ "embed"
	"encoding/json"
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
	"github.com/go-pkgz/rest"
	"github.com/parMaster/mcache"
	"github.com/parMaster/rpid/config"
	"github.com/parMaster/rpid/storage"
	"github.com/parMaster/rpid/storage/model"
	"github.com/parMaster/rpid/web"
	flags "github.com/umputun/go-flags"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

type historical map[string][]int

type Worker struct {
	config  config.Parameters
	revs    int // persistent revs counter
	data    historical
	i2cBus  i2c.BusCloser
	modules Modules
	mx      sync.Mutex
	store   storage.Storer
	cache   mcache.Cacher
	ctx     context.Context
}

func NewWorker(config *config.Parameters) *Worker {
	data := historical{
		// CPU Temperature in milliCentigrades
		"t":    {}, // momentary temp
		"temp": {}, // temp history by minute
	}

	w := &Worker{
		config: *config,
		data:   data,
		cache:  mcache.NewCache(),
	}

	return w
}

func (w *Worker) Run(ctx context.Context) error {
	var err error
	w.ctx = ctx

	if err = storage.Load(ctx, w.config.Storage, &w.store); w.config.Storage.Type != "" && err != nil {
		log.Printf("[ERROR] failed to load storage: %v", err)
	}

	// Load peripheral drivers
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	if w.config.Modules.I2C != "" {
		w.i2cBus, err = i2creg.Open(w.config.Modules.I2C)
		if err != nil {
			log.Printf("[ERROR] failed to open I²C: %v", err)
			return err
		}
	}

	w.loadModules()
	log.Printf("[DEBUG] Loaded modules: %s", w.modules)

	go w.controlFan(ctx)
	go w.startTach(ctx)

	go w.logEverySecond(ctx)
	go w.logEveryMinute(ctx)
	go w.startServer(ctx)

	log.Printf("Service started. Fan tach on %s, trigger on %s, listening to \"%s\"", w.config.Fan.TachPin, w.config.Fan.ControlPin, w.config.Server.Listen)
	log.Printf("Temps cfg: low=%d˚C, high=%d˚C", w.config.Fan.Low, w.config.Fan.High)
	if w.store != nil {
		log.Printf("Storage: %s, %s", w.config.Storage.Type, w.config.Storage.Path)
	}

	<-ctx.Done()
	time.Sleep(2 * time.Second) // wait 2 secs till tach timeout (1 sec) hits
	if w.i2cBus != nil {
		log.Println("[DEBUG] Closing I²C Bus on exit")
		if err := w.i2cBus.Close(); err != nil {
			log.Printf("[ERROR] Closing I²C: %e", err)
			return err
		}
	}
	return nil
}

func (w *Worker) setFanState(fanControl gpio.PinIO, state bool) error {
	if err := fanControl.Out(gpio.Level(state)); err != nil {
		log.Printf("[ERROR] Changing fan state (%v): %e", state, err)
		return err
	}
	log.Printf("[DEBUG] Fan set to %v", gpio.Level(state))
	return nil
}

func (w *Worker) controlFan(ctx context.Context) {
	if w.config.Fan.ControlPin == "" {
		log.Println("[INFO] No fan ControlPin defined, skipping fan control")
		return
	}
	fanControl := gpioreg.ByName(w.config.Fan.ControlPin)
	if fanControl == nil {
		log.Printf("[ERROR] Failed to find %s", fanControl)
	}
	time.Sleep(1 * time.Second)

	tempHigh := w.config.Fan.High * 1000 // fan   activation temperature m˚C
	tempLow := w.config.Fan.Low * 1000   // fan DEactivation temperature m˚C
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("[DEBUG] Leaving the fan ON is always safer")
			fanControl.Out(gpio.High)
			fanControl.Halt()
			return
		case <-ticker.C:
		}

		w.mx.Lock()
		ma10sec, ma30sec, ma1min, ma3min := 0, 0, last(w.data["temp"]), 0
		if len(w.data["t"]) >= 10 {
			ma10sec = avg(w.data["t"][max(0, len(w.data["t"])-9) : len(w.data["t"])-1])
		}
		log.Printf("[DEBUG] 10 seconds moving average: %d", ma10sec)

		if len(w.data["t"]) >= 30 {
			ma30sec = avg(w.data["t"][max(0, len(w.data["t"])-29) : len(w.data["t"])-1])
		}
		log.Printf("[DEBUG] 30 seconds moving average: %d", ma30sec)

		log.Printf("[DEBUG] 1 minute moving average: %d", ma1min)

		if len(w.data["temp"]) >= 3 {
			ma3min = avg(w.data["temp"][max(0, len(w.data["temp"])-2) : len(w.data["temp"])-1])
		}
		log.Printf("[DEBUG] 3 minutes moving average: %d", ma3min)
		w.mx.Unlock()

		// Fan activation conditions
		if ma10sec > tempHigh+10000 || // Sudden spike
			ma30sec > tempHigh+5000 || // Fast rise
			ma1min > tempHigh || // High temperature
			ma1min == 0 || // No data
			ma3min == 0 { //  No data

			w.setFanState(fanControl, true)
			continue
		}

		// Deactivate otherwise
		if ma3min < tempLow || // Lower than low for 3 minutes
			ma1min < tempLow-1000 || // Low enough
			ma30sec < tempLow-2000 || // Fast decline
			ma10sec < tempLow-4000 { // Sudden drop
			w.setFanState(fanControl, false)
		}
	}
}

func (w *Worker) startTach(ctx context.Context) {
	if w.config.Fan.TachPin == "" {
		log.Println("[INFO] No tachymeter configured")
		return
	}
	w.data["revs"] = []int{}
	w.data["rpm"] = []int{}
	var tach gpio.PinIn = gpioreg.ByName(w.config.Fan.TachPin)
	if tach == nil {
		log.Fatalf("Failed to find %s", w.config.Fan.TachPin)
	}

	// Set pin as input, with an internal pull-up resistor:
	if err := tach.In(gpio.PullUp, gpio.RisingEdge); err != nil {
		log.Fatal(err)
	}

	// Count every rev or exit
	for {
		select {
		case <-ctx.Done():
			log.Println("[DEBUG] Halting tachymeter")
			if err := tach.Halt(); err != nil {
				log.Printf("[ERROR] Halting tachymeter: %e", err)
			}
			return
		default:
		}
		if tach.WaitForEdge(time.Second) {
			w.revs++
		}
	}
}

func (w *Worker) startServer(ctx context.Context) {
	httpServer := &http.Server{
		Addr:              w.config.Server.Listen,
		Handler:           w.router(),
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Second,
	}

	httpServer.ListenAndServe()

	<-ctx.Done()
	log.Printf("[INFO] Terminating http server")

	if err := httpServer.Close(); err != nil {
		log.Printf("[ERROR] failed to close http server, %v", err)
	}
}

func (w *Worker) router() http.Handler {
	router := chi.NewRouter()
	router.Use(rest.Throttle(5))

	router.Get("/status", func(rw http.ResponseWriter, r *http.Request) {
		w.mx.Lock()
		resp := map[string]int{
			"temp": last(w.data["temp"]) / 1000,
			"rpm":  last(w.data["rpm"]),
		}
		w.mx.Unlock()

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(resp)
	})

	router.Get("/fullData", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Access-Control-Allow-Origin", "*")
		out := w.getFullData()
		json.NewEncoder(rw).Encode(out)
	})

	router.Get("/charts", func(rw http.ResponseWriter, r *http.Request) {
		w.responseWithFile("web/chart.html", rw)
	})
	router.Get("/web/chart_tpl.min.js", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/javascript")
		w.responseWithFile("web/chart_tpl.min.js", rw)
	})
	router.Get("/web/plotly.min.js", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/javascript")
		w.responseWithFile("web/plotly.min.js", rw)
	})
	router.Get("/view", func(rw http.ResponseWriter, r *http.Request) {
		w.responseWithFile("web/view.html", rw)
	})

	router.Get("/viewData/{module}", func(rw http.ResponseWriter, r *http.Request) {
		if w.store == nil {
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		module := chi.URLParam(r, "module")
		if module == "" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Header().Set("Access-Control-Allow-Origin", "*")

		out, err := w.cache.Get(module)
		if err != nil {
			out, err = w.store.View(w.ctx, module)
			if err != nil {
				log.Printf("[ERROR] Failed to get view: %v", err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.cache.Set(module, out, 60)
		}
		json.NewEncoder(rw).Encode(out)
	})

	return router
}

func (w *Worker) responseWithFile(file string, rw http.ResponseWriter) error {
	var html []byte
	var err error
	if w.config.Server.Dbg {
		html, err = os.ReadFile(file)
	} else {
		file = file[4:] // cut off web/ prefix
		html, err = web.WebAssets.ReadFile(file)
	}
	if err != nil {
		log.Printf("[ERROR] failed to read %s, %v", file, err)
		rw.WriteHeader(http.StatusInternalServerError)
		return err
	}
	rw.Write(html)
	return nil
}

func (w *Worker) getFullData() interface{} {
	w.mx.Lock()
	defer w.mx.Unlock()

	var out struct {
		Data    historical
		Dates   []string
		Modules map[string]interface{}
	}

	// dates are not stored but generated on the fly
	out.Data = w.data
	out.Data["revs"] = []int{}
	out.Dates = []string{}
	now := time.Now()
	for i := len(out.Data["temp"]); i > 0; i-- {
		out.Dates = append(out.Dates, now.Add(-1*time.Minute*time.Duration(i)).Format("2006-01-02 15:04"))
	}

	out.Modules = make(map[string]interface{})
	for _, m := range w.modules {
		data, err := m.Report()
		if err != nil {
			log.Printf("[ERROR] %s: %v", m.Name(), err)
		}
		out.Modules[m.Name()] = data
	}

	return out
}

func (w *Worker) logEverySecond(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	var temp int
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		temp = 0
		// Current temperature as reported by thermal zone (sensor), millidegree Celsius
		// https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-class-thermal
		sysTemp, err := os.ReadFile("/sys/class/thermal/thermal_zone0/temp")
		if err != nil {
			log.Printf("[ERROR] Can't read temperature file: %e", err)
		} else {
			temp, err = strconv.Atoi(string(sysTemp[0 : len(sysTemp)-1]))
			if err != nil {
				log.Printf("[ERROR] Converting temp data: %e", err)
			}
		}

		if w.config.Server.Dbg {
			log.Printf("[DEBUG] Temp: %d m˚C | Fan RPS/RPM: %d/%d\r\n", temp, w.revs, w.revs*60)
		}

		w.mx.Lock()
		w.data["revs"] = append(w.data["revs"], w.revs*60)
		w.revs = 0
		w.data["t"] = append(w.data["t"], temp)
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

		w.mx.Lock()
		if w.config.Fan.TachPin != "" {
			w.data["rpm"] = append(w.data["rpm"], avg(w.data["revs"]))
		}
		w.data["revs"] = []int{}
		w.data["temp"] = append(w.data["temp"], avg(w.data["t"][max(0, len(w.data["t"])-60):len(w.data["t"])-1]))

		// "scrolling" temperature history, leave only last 60-120 seconds
		if len(w.data["t"]) > 100 {
			w.data["t"] = w.data["t"][len(w.data["t"])-60 : len(w.data["t"])-1]
		}

		log.Printf("CPU: %d m˚C\r\n", last(w.data["temp"]))
		log.Printf("Fan: %d rpm\r\n", last(w.data["rpm"]))

		if w.store != nil {
			w.store.Write(ctx, model.Data{Module: "main", Topic: "temp", Value: fmt.Sprint(last(w.data["temp"]))})
			w.store.Write(ctx, model.Data{Module: "main", Topic: "rpm", Value: fmt.Sprint(last(w.data["rpm"]))})
		}

		for _, m := range w.modules {
			err := m.Collect(ctx)
			if err != nil {
				log.Printf("[ERROR] %s: %v", m.Name(), err)
			}
		}

		w.mx.Unlock()
	}
}

// Load modules
func (w *Worker) loadModules() (names []string) {

	if w.config.Modules.System.Enabled {
		sys, err := LoadSystemReporter(w.config.Modules.System, w.store, w.config.Server.Dbg)
		if err != nil {
			log.Printf("%e", err)
		} else {
			w.modules = append(w.modules, sys)
			names = append(names, sys.Name())
		}
	}

	if w.config.Modules.BMP280.Enabled {
		modbmp280, err := LoadBmp280Reporter(w.config.Modules.BMP280, w.i2cBus, w.store)
		if err != nil {
			log.Printf("%e", err)
		} else {
			w.modules = append(w.modules, modbmp280)
			names = append(names, modbmp280.Name())
		}
	}

	if w.config.Modules.HTU21.Enabled {
		modhtu21, err := LoadHtu21Reporter(w.config.Modules.HTU21, w.i2cBus)
		if err != nil {
			log.Printf("%e", err)
		} else {
			w.modules = append(w.modules, modhtu21)
			names = append(names, modhtu21.Name())
		}
	}

	if w.config.Modules.Smc768.Enabled {
		modsmc768, err := LoadSmc768Reporter(w.config.Modules.Smc768, w.store, w.config.Server.Dbg)
		if err != nil {
			log.Printf("%e", err)
		} else {
			w.modules = append(w.modules, modsmc768)
			names = append(names, modsmc768.Name())
		}
	}

	return
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
	Config string `long:"config" env:"CONFIG" default:"config.yml" description:"yaml config file name"`
	Dbg    bool   `long:"dbg" env:"DEBUG" description:"show debug info"`
}

func main() {
	// Parsing cmd parameters
	var opts Options
	p := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		p.WriteHelp(os.Stderr)
		os.Exit(2)
	}

	var conf *config.Parameters
	if opts.Config != "" {
		var err error
		conf, err = config.NewConfig(opts.Config)
		if err != nil {
			log.Fatalf("[ERROR] can't load config, %s", err)
		}
		conf.Server.Dbg = opts.Dbg
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
		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("Shutdown signal received\n*********************************")
		cancel()
	}()

	defer func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic: %+v", x)
		}
	}()

	NewWorker(conf).Run(ctx)
}
