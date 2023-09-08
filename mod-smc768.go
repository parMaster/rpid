package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/parMaster/rpid/config"
	"github.com/parMaster/rpid/storage"
	"github.com/parMaster/rpid/storage/model"
)

// Sensors is a list of the sensors we want to monitor
var Sensors = []string{
	// SMC sensors
	"TA0V", // TA0V is the ambient virtual temperature
	"TC0C", // TC0C is the CPU core temperature
	"TCGC", // TCGC is the GPU core temperature
	"TH1A", // TH1A is the Drive 0 temperature
	"TM0P", // TM0P is the memory temperature
	"TS0V", // TS0V is the enclosure virtual temperature
	"TW0P", // TW0P is the wireless module temperature
	"Te0T", // Te0T is the enclosure temperature
	"Tm0P", // Tm0P is the memory temperature
	// Other sensors
	"Exhaust",      //Exhaust fan speed in RPM
	"ThrottleTime", // Core throttle time in milliseconds
}

type Smc768Response map[string]string

type Smc768Reporter struct {
	data  Smc768Response
	dbg   bool
	mx    sync.Mutex
	store storage.Storer
}

func LoadSmc768Reporter(cfg config.Smc768, store storage.Storer, dbg bool) (*Smc768Reporter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Smc768Reporter is not enabled")
	}

	if store != nil {
		log.Printf("[DEBUG] Smc768Reporter: using storage (%T)", store)
	}

	return &Smc768Reporter{
		dbg:   dbg,
		store: store,
		data:  make(map[string]string),
	}, nil
}

func (r *Smc768Reporter) Name() string {
	return "smc768"
}

func (r *Smc768Reporter) Collect(ctx context.Context) (err error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.data = r.ReadSMC768()
	if err != nil {
		return errors.Join(err, fmt.Errorf("failed to get load avg: %v", err))
	}

	if r.store != nil {
		for _, label := range Sensors {
			err := r.store.Write(ctx, model.Data{Module: r.Name(), Topic: label, Value: r.data[label]})
			if err != nil {
				return errors.Join(err, fmt.Errorf("failed to write to storage: %v", err))
			}
		}
	}
	return err
}

func (r *Smc768Reporter) Report() (interface{}, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.data, nil
}

func (r *Smc768Reporter) ReadSMC768() Smc768Response {

	data := make(Smc768Response)

	// Read the data from the SMC768 and store it in the data map
	for i := 1; i <= 60; i++ {
		value := ReadInput(fmt.Sprintf("/sys/devices/platform/applesmc.768/temp%d_input", i))
		label := ReadInput(fmt.Sprintf("/sys/devices/platform/applesmc.768/temp%d_label", i))
		if slices.Contains(Sensors, label) {
			data[label] = value
		}
	}

	data["Exhaust"] = ReadInput("/sys/devices/platform/applesmc.768/fan1_input")
	data["ThrottleTime"] = ReadInput("/sys/devices/system/cpu/cpu0/thermal_throttle/core_throttle_total_time_ms")

	if r.dbg {
		log.Printf("[DEBUG] Smc768Reporter: data:")
		for k, v := range data {
			fmt.Printf("%s: %s: %s\n", k, v, smc[k])
		}
	}

	return data
}

func ReadInput(file string) (v string) {

	value, err := os.ReadFile(file)
	if err != nil {
		fmt.Println(err)
	}

	v = strings.TrimSpace(string(value))

	return
}

// SMC keys from https://github.com/acidanthera/VirtualSMC/blob/master/Docs/SMCSensorKeys.txt
// parsed, sorted, deduped, and formatted as a Go map
var smc = map[string]string{
	"B0AC": "Battery Actual current. (mA) (B0AC)",
	"B0AV": "Battery Actual Voltage. (V) (B0AV)",
	"B0FC": "Battery full charge capacity. (mAh) (B0FC)",
	"B0RM": "Battery remain capacity. (mAh) (B0RM)",
	"CHBI": "Battery charge current. (mAh) (CHBI) // usrsse2: should be mA, not mAh",
	"F0Ac": "Fan0 Actual RPM(F0Ac)",
	"F1Ac": "Fan1 Actual RPM(F1Ac)",
	"FRD0": "gasi16TmFastRawData array entry 0 (FRD0)",
	"FRD1": "gasi16TmFastRawData array entry 1 (FRD1)",
	"FRD2": "gasi16TmFastRawData array entry 2 (FRD2)",
	"FRD3": "gasi16TmFastRawData array entry 3 (FRD3)",
	"I50R": "5V lowside current (Amps) (I50R)",
	"IAPC": "Airport Current (DEBUG) (Amps) (IAPC)",
	"IB0R": "Battery Discharge Current MLB . (Amps) (IB0R)",
	"IBAC": "Battery Current (IBAC)",
	"IBLC": "Backlight Current (DEBUG) (Amps) (IBLC)",
	"IC0C": "CPU Core current. (Amps) (IC0C)",
	"IC0G": "CPU AXG low-side current (Amps) (IC0G)",
	"IC0I": "CPU I/O high-side current (Amps) (IC0I)",
	"IC0M": "CPU Mem load-side current (Amps) (IC0M)",
	"IC0R": "CPU High Side Current (Amps) (IC0R)",
	"IC0S": "CPU VSA low side current (Amps) (IC0S)",
	"IC1C": "VCCIO 1.05V S0 current. (Amps) (IC1C)",
	"IC1R": "CPU High Side Current from EMC1704 (Amps) (IC1R)",
	"IC2C": "VCCSA S0 current. (Amps) (IC2C)",
	"IC3C": "CPU DDR Current (Amps) (IC3C)",
	"ICMC": "S2 Camera Current (Amps) (ICMC)",
	"ICS0": "CPU Core current. (Amps) (ICS0)",
	"ICS1": "CPU current from CPU IMON. (Amps) (ICS1)",
	"ICTR": "CPU Riser 12V high side current (Amps) (ICTR)",
	"ID0R": "DC In current (Amps) (ID0R)",
	"ID1R": "CPU 1V5 current (Amps) (ID1R)",
	"ID2E": "Power Supply 12V current filtered (Amps) (ID2E)",
	"ID2F": "Power Supply 12V current filtered and adjusted (ID2F)",
	"ID2I": "Power Supply 12V current filter input S7.8 (Amps) (ID2I)",
	"ID2J": "Power Supply 12V current max error (Amps) (ID2J)",
	"ID2L": "Power Supply 12V current dynamic filter coeff (Seconds) (ID2L)",
	"ID2R": "Power Supply 12V current (Amps) (ID2R)",
	"ID2T": "Power Supply 12V current trend (Amps) (ID2T)",
	"IG0C": "GPU Core low-side current (Amps) (IG0C)",
	"IG0F": "GPU FB high-side current (Amps) (IG0F)",
	"IG0I": "GPU VDDCI load-side current (Amps) (IG0I)",
	"IG0R": "Ext GPU High Side Current. (Amps) (IG0R)",
	"IG0S": "GPU A VDDCI low side current (Amps) (IG0S)",
	"IG0U": "GPU Uncore high-side current (Amps) (IG0U)",
	"IG1A": "GPU Aux input-side current (Amps) (IG1A)",
	"IG1C": "GPU B Core low side current (Amps) (IG1C)",
	"IG1F": "GPU FB input-side current (Amps) (IG1F)",
	"IG1R": "GFX B Riser 12V high side current (Amps) (IG1R)",
	"IG1S": "GPU B VDDCI low side current (Amps) (IG1S)",
	"IG2C": "Ext GPU 1.0V Current (DEBUG) (Amps) (IG2C)",
	"IG3C": "Ext GPU VRAM & I/O current Current (DEBUG) (Amps) (IG3C)",
	"IH02": "Drive 0 12V load-side current (Amps) (IH02)",
	"IH05": "HDD 5V low-side current (Amps) (IH05)",
	"IH0R": "HDD1 lowside current (Amps) (IH0R)",
	"IH1R": "SSD lowside current (Amps) (IH1R)",
	"IHDC": "HDD Current (DEBUG) (Amps) (IHDC)",
	"IHSC": "Thunderbolt Current (Amps) (IHSC)",
	"IHSP": "Thunderbolt Current (Amps) (IHSP)",
	"II0R": "I/O Board 12V high side current (Amps) (II0R)",
	"ILDC": "LCD panel Current (Amps) (ILDC)",
	"IM0C": "DDR3 Current (Amps) (IM0C)",
	"IM0R": "DDR 1.5V (Memory) current. (Amps) (IM0R)",
	"IM1C": "DDR 1.5V Current (DEBUG) (Amps) (IM1C)",
	"IM2C": "LPDDR current 1.8V. (Amps) (IM2C)",
	"IM3C": "DDR S3 current. (Amps) (IM3C)",
	"IMTR": "Memory 1V5 high side current (Amps) (IMTR)",
	"IN0C": "Internal GFX Core current. (Amps) (IN0C)",
	"IN0R": "GFX Vcore current (Amps) (IN0R)",
	"IN1C": "MCP DDR Current(SPS! force bit 11) (Amps) (IN1C)",
	"IN1R": "PCH 1V05 current (Amps) (IN1R)",
	"IO0R": "Other High Side Current (Amps) (IO0R)",
	"IO3R": "Other 3.3V High Side current. (Amps) (IO3R)",
	"IO5R": "Battery current BMON. (Amps) (IO5R)",
	"IODC": "ODD Current (DEBUG) (Amps) (IODC)",
	"IP0R": "Battery Current Discrete. (Amps) (IP0R)",
	"IPB1": "Discrete battery current. (Amps) (IPB1)",
	"IPBR": "PBus on battery current. (Amps) (IPBR)",
	"IR0C": "S0 5.0V Current (DEBUG) (Amps) (IR0C)",
	"IR1C": "S3 3.30V Current (DEBUG) (Amps) (IR1C)",
	"IR1R": "PCH/GPU/TBT 1.05V high-side current (Amps) (IR1R)",
	"IR2C": "S3 5.0V Current (DEBUG) (Amps) (IR2C)",
	"IR3C": "S5 3.30V Current (DEBUG) (Amps) (IR3C)",
	"IR5C": "3.3V S5 current. (Amps) (IR5C)",
	"IS2C": "S2 Camera Current (Amps) (IS2C)",
	"ISBC": "PCH core Current (DEBUG) (Amps) (ISBC)",
	"ISDC": "SSD Current (Amps) (ISDC)",
	"ITPC": "T25 - Track Pad Current (Amps) (ITPC)",
	"IW0R": "Wifi load-side current (Amps) (IW0R)",
	"IZAP": "Airport Current (DEBUG) (Amps) (IZAP)",
	"IZBL": "Backlight Current(IZBL)",
	"IZDM": "DIMM Current(IZDM)",
	"IZHD": "HDD Current(IZHD)",
	"IZOD": "ODD Current(IZOD)",
	"IZOP": "Airport Current(IZOP)",
	"Ic0R": "CSreg Current (SPS! force bit 12) (Amps) (Ic0R)",
	"MSLT": "HP Power Throttle (MSLT)",
	"MST1": "Audio ScratchPad register (MST1)",
	"MST2": "Audio ScratchPad2 register (MST2)",
	"MSTc": "CPU Plimit (MSTc)",
	"MSTf": "Forced Idle Limit (MSTf)",
	"MSTg": "GPU Plimit (MSTg)",
	"MSTi": "2nd GPU Plimit (MSTi)",
	"MSTm": "MEM Plimit (MSTm)",
	"P50R": "5V lowside power (I50R * V50R) (Watts) (P50R)",
	"PAPC": "WLAN lowside power (IAPC * 3.3V) (Watts) (PAPC)",
	"PB0R": "Battery Discharge Power MLB (IB0R * VP0R) (Watts) (PB0R)",
	"PB1R": "Discrete BMON on battery power (Watts) (PB1R)",
	"PBLC": "Backlight Input Power (DEBUG) (Watts) (PBLC)",
	"PC0C": "CPU Core low side power (Watts) (PC0C)",
	"PC0G": "CPU AXG low-side power (Watts) (PC0G)",
	"PC0I": "CPU I/O high-side power (Watts) (PC0I)",
	"PC0M": "CPU Mem low-side power (Watts) (PC0M)",
	"PC0R": "CPU Computing High Side Power (Watts) (PC0R)",
	"PC0S": "CPU VSA low side power (Watts) (PC0S)",
	"PC1C": "CPU VCCIO 1.05V S0 Power (Watts) (PC1C)",
	"PC1R": "CPU High Side Power from EMC1704 (Watts) (PC1R)",
	"PC2C": "CPU VCCSA Power (Watts) (PC2C)",
	"PC3C": "CPU DDR Power (Watts) (PC3C)",
	"PCLT": "CPU Package Total Power (SMC Load Sensors) (Watts) (PCLT)",
	"PCMC": "S2 Camera Power (Watts) (PCMC)",
	"PCPC": "CPU package core power (PECI) (Watts) (PCPC)",
	"PCPG": "CPU package Gfx power (PECI) (Watts) (PCPG)",
	"PCPR": "CPU package total power (SMC) (PCPR)",
	"PCPT": "CPU package total power (PECI) (Watts) (PCPT)",
	"PCS0": "Chipset lowside Power (Watts) (PCS0)",
	"PCS1": "Chipset IMon Power (Watts) (PCS1)",
	"PCTR": "CPU Riser total power (Watts) (PCTR)",
	"PD0D": "Power Supply Power Trend (Watts) (PD0D)",
	"PD0E": "Power supply filtered power (Watts) (PD0E)",
	"PD0F": "Power supply power filtered then adjusted (Watts) (PD0F)",
	"PD0J": "Power supply max error filtered output (Watts) (PD0J)",
	"PD0K": "Power supply filter coefficient (Watts) (PD0K)",
	"PD0R": "DC-In rail Power (Watts) (PD0R)",
	"PD1R": "CPU 1V5 power (ID1R * 1.5V) (Watts) (PD1R)",
	"PD2R": "Power Supply 12V power (Watts) (PD2R)",
	"PDTR": "DC-In total power (Watts) (PDTR)",
	"PG0A": "GPU Aux load-side power (= x% of PG1A) (Watts) (PG0A)",
	"PG0C": "GPU A Core low side power (Watts) (PG0C)",
	"PG0F": "GPU Frame Buf high-side power (Watts) (PG0F)",
	"PG0I": "GPU VDDCI load-side power (Watts) (PG0I)",
	"PG0M": "GPU Mem Ctl load-side power (= 30% of PG1F * efficiency) (Watts) (PG0M)",
	"PG0R": "GFX A total power (Watts) (PG0R)",
	"PG0S": "GPU A VDDCI low side power (Watts) (PG0S)",
	"PG0U": "GPU Uncore high-side power (Watts) (PG0U)",
	"PG0V": "GPU VRAM load-side power (= 70% of PG1F * efficiency) (Watts) (PG0V)",
	"PG1A": "GPU Aux input-side power (Watts) (PG1A)",
	"PG1C": "GPU B Core low side power (Watts) (PG1C)",
	"PG1F": "GPU Frame Buf input-side power (Watts) (PG1F)",
	"PG1R": "GFX B total power (Watts) (PG1R)",
	"PG1S": "GPU B VDDCI low side power (Watts) (PG1S)",
	"PG2C": "Ext GPU 1.8V Power (Watts) (PG2C)",
	"PG3C": "Ext GPU 1.8V Power (DEBUG) (Watts) (PG3C)",
	"PGPR": "GPU package total load-side power (SMC) (PGPR)",
	"PGTR": "GPU total high-side power (PG0R + PG1R) (Watts) (PGTR)",
	"PH02": "Drive 0 12V low-side power (Watts) (PH02)",
	"PH05": "Drive 0 5V low-side power (Watts) (PH05)",
	"PH0F": "Drive 0 filtered total load-side power (PH02 + PH05) * 2 (Watts) (PH0F)",
	"PH0R": "SSD 3V3 low side power (Watts) (PH0R)",
	"PH1F": "Drive 1 filtered load-side power * 3 (Watts) (PH1F)",
	"PH1R": "Drive 1 low-side power (Watts) (PH1R)",
	"PH2R": "Drive 2 power (IH2R * 3.3V) (Watts) (PH2R)",
	"PHDC": "HDD Power (DEBUG) (Watts) (PHDC)",
	"PHPC": "HP Power (PHPC)",
	"PHSC": "Thunderbolt Power (Watts) (PHSC)",
	"PHSP": "T29 Power (Watts) (PHSP)",
	"PI0R": "I/O Board total power (Watts) (PI0R)",
	"PLDC": "LCD Driver Power (Watts) (PLDC)",
	"PM0C": "DDR Power 1.2V CPU & MEM (Watts) (PM0C)",
	"PM0R": "DIMM low-side power (Watts) (PM0R)",
	"PM0f": "DIMM load-side power / SQR(Fan RPM) (PM0f)",
	"PM1C": "DDR Power 1.2V CPU (Watts) (PM1C)",
	"PM2C": "DDR Power 1.8V CPU (Watts) (PM2C)",
	"PM3C": "DDR S3 Power. (Watts) (PM3C)",
	"PMTR": "Memory total power (Watts) (PMTR)",
	"PN0C": "AXG core Power (Watts) (PN0C)",
	"PN0R": "GFX Vcore power (IN0R * VN0R) (Watts) (PN0R)",
	"PN1C": "Average MCP memory Power (IN1C * 1.35V) (SPS! force bit 24) (Watts) (PN1C)",
	"PN1R": "PCH low-side power (Watts) (PN1R)",
	"PO0R": "Other High Side Power (Watts) (PO0R)",
	"PO3R": "Other 3.3V High Side Power (Watts) (PO3R)",
	"PO5R": "Other 5V High Side Power (Watts) (PO5R)",
	"PP0R": "PBus rail Power (Watts) (PP0R)",
	"PPBR": "PBus on battery power (Watts) (PPBR)",
	"PR0C": "3.3V S0 power (Watts) (PR0C)",
	"PR0R": "S0 5.0V Power (DEBUG) (Watts) (PR0R)",
	"PR1C": "3.3V S3. (Amps) (PR1C)",
	"PR1R": "1.05V high-side power (Watts) (PR1R)",
	"PR2R": "S3 5.0V Power (DEBUG) (Watts) (PR2R)",
	"PR3C": "3.3V in S0 Power(Watts) (PR3C)",
	"PR3R": "S5 3.30V Power (DEBUG) (Watts) (PR3R)",
	"PR5C": "5VV in S0 Power(Watts) (PR5C)",
	"PS2C": "S2 Camera power (Watts) (PS2C)",
	"PSBC": "PCH core Power (DEBUG) (Watts) (PSBC)",
	"PSDC": "SSD Fixed 3.3V Power (Watts) (PSDC)",
	"PSTR": "System Total Power Consumed (Delayed 1 Second) (Watts) (PSTR)",
	"PTHC": "HP Power Target (PTHC)",
	"PTPC": "T101 Accuator Power (Watts) (PTPC)",
	"PTPR": "T25 - Track Pad Power (Watts) (PTPR)",
	"PW0R": "Wifi low-side power (Watts) (PW0R)",
	"PZ0E": "Zone 0 Target Power (PZ0E)",
	"PZ0F": "Zone 0 Filtered Power (PZ0F)",
	"PZ0G": "Zone 0 average power (Watts) (PZ0G)",
	"PZ0T": "Zone 0 Abstract Throttle (PZ0T)",
	"PZ1E": "Zone 1 Target Power (PZ1E)",
	"PZ1F": "Zone 1 Filtered Power (PZ1F)",
	"PZ1G": "Zone 1 average power (Watts) (PZ1G)",
	"PZ1T": "Zone 1 Abstract Throttle (PZ1T)",
	"PZ2E": "Zone 1 Target Power (PZ2E)",
	"PZ2F": "Zone 1 Filtered Power (PZ2F)",
	"PZ2G": "Zone 2 average power (Watts) (PZ2G)",
	"PZ2T": "Zone 2 Abstract Throttle (PZ2T)",
	"PZ3E": "Zone 1 Target Power (PZ3E)",
	"PZ3F": "Zone 1 Filtered Power (PZ3F)",
	"PZ3G": "Zone 3 average power (Watts) (PZ3G)",
	"PZ3T": "Zone 3 Abstract Throttle (PZ3T)",
	"PZ4E": "Zone 4 Target Power (PZ4E)",
	"PZ4F": "Zone 4 Filtered Power (PZ4F)",
	"PZ4G": "Zone 4 average power (Watts) (PZ4G)",
	"PZ4T": "Zone 4 Abstract Throttle (PZ4T)",
	"PZ5E": "Zone 5 Target Power (PZ5E)",
	"PZ5F": "Zone 5 Filtered Power (PZ5F)",
	"PZ5G": "Zone 5 average power (Watts) (PZ5G)",
	"PZ5T": "Zone 5 Abstract Throttle (PZ5T)",
	"PZAP": "WIFI Power (IZAP * 3.3V) (SPS! force bit 27) (Watts) (PZAP)",
	"PZBL": "LCD Backlight Power (IZBL * VP0R) (SPS! force bit 26) (Watts) (PZBL)",
	"PZDM": "Memory Power (IZDM * 1.35V) (SPS! force bit 29) (Watts) (PZDM)",
	"PZHD": "SSD Power (IZHD * 3.3V) (SPS! force bit 28) (Watts) (PZHD)",
	"PZOD": "ODD Power(PZOD)",
	"Pc0R": "Average main chipset Power (Ic0R * VP0R) (SPS! force bit 25) (Watts) (Pc0R)",
	"TA0P": "Ambient 1 temp (DegC) (TA0P)",
	"TA0V": "Ambient virtual temp (DegC) (TA0V)",
	"TA0p": "Ambient 1 raw temp (DegC) (TA0p)",
	"TA1P": "Ambient 2 temp (DegC) (TA1P)",
	"TA1p": "Ambient 2 raw temp (DegC) (TA1p)",
	"TA2P": "Ambient 3 temp (DegC) (TA2P)",
	"TA2p": "Ambient 3 raw temp (DegC) (TA2p)",
	"TB0T": "Battery temp (SBS! force bit 0) (DegC) (TB0T)",
	"TB1T": "Battery Thermistor Temp 1 (TB1T)",
	"TB2T": "Battery Thermistor Temp 2 (TB2T)",
	"TB3T": "Battery Temp(TB3T)",
	"TBXT": "Battery temp (Same as TB0T) (DegC) (TBXT)",
	"TC0C": "CPU Core PECI Temp(TC0C)",
	"TC0D": "CPU die temp (DegC) (TC0D)",
	"TC0E": "CPU PECI Die filtered temp for fan control (DegC) (TC0E)",
	"TC0F": "CPU PECI Die filtered and adjusted temp for power control (DegC) (TC0F)",
	"TC0J": "CPU PECI die temp max error filtered output used in TC0F=TC0E+TC0G (DegC) (TC0J)",
	"TC0L": "CPU PECI Die dynamic filter coeff (Seconds) (TC0L)",
	"TC0P": "CPU proximity temp (filtered) (DegC) (TC0P)",
	"TC0T": "CPU PECI Die temp Trend (DegC) (TC0T)",
	"TC0c": "CPU Core 0 absolute raw temp (PECI) (DegC) (TC0c)",
	"TC0d": "CPU die raw temp (DegC) (TC0d)",
	"TC0p": "CPU proximity raw temp (DegC) (TC0p)",
	"TC1C": "Max of CPU Core vs. Mem Ctlr/IG PECI temps (DegC) (SIS! force bit 5) (TC1C)",
	"TC1P": "CPU VR Proximity cooked temp (DegC) (TC1P)",
	"TC1c": "CPU Core 1 absolute raw temp (PECI) (DegC) (TC1c)",
	"TC2C": "CPU Core 2 temp (PECI) (DegC) (TC2C)",
	"TC2c": "CPU Core 2 absolute raw temp (PECI) (DegC) (TC2c)",
	"TC3C": "CPU Core 3 temp (PECI) (DegC) (TC3C)",
	"TC3c": "CPU Core 3 absolute raw temp (PECI) (DegC) (TC3c)",
	"TC4C": "CPU Core 3 temp (PECI) (DegC) (TC4C)",
	"TCFC": "CPU PECI die temp filter coeff (TCFC)",
	"TCGC": "CPU Gfx Core temp (PECI) (DegC) (TCGC)",
	"TCGc": "CPU Gfx absolute raw temp (PECI) (DegC) (TCGc)",
	"TCHP": "Charger proximity temp (DegC) (TCHP)",
	"TCMX": "Max PECI reported temp (DegC) (TCMX)",
	"TCMc": "CPU eDRAM Hot Indicator(PECI) (DegC) (TCMc)",
	"TCSA": "CPU System Agent Core temp (PECI) (DegC) (TCSA)",
	"TCSC": "CPU System Agent Core temp (PECI) (DegC) (TCSC)",
	"TCSc": "CPU System Agent Core absolute raw temp (PECI) (DegC) (TCSc)",
	"TCTD": "CPU PECI die temp Trend (SIS! force bit 5) (DegC) (TCTD)",
	"TCXC": "CPU Max Package Core absolute cooked temp (PECI) (DegC) (TCXC)",
	"TCXR": "CPU Max Package Core relative cooked temp (PECI) (DegC) (TCXR)",
	"TCXc": "CPU Max Package Core absolute raw temp (PECI) (DegC) (TCXc)",
	"TCXr": "CPU Max Package Core relative raw temp (PECI) (DegC) (TCXr)",
	"TG0D": "GPU Die cooked temp (DegC) (TG0D)",
	"TG0E": "GPU Die filtered temp (DegC) (TG0E)",
	"TG0F": "GPU Die filtered and adjusted temp for fan/power control (DegC) (TG0F)",
	"TG0H": "Heat Pipe Temp(TG0H)",
	"TG0J": "GPU Die temp max error used in TG0F=TG0E+ (DegC) (TG0J)",
	"TG0L": "GPU Die dynamic filter coeff (Seconds) (TG0L)",
	"TG0M": "GFX memory temp (DegC) (TG0M)",
	"TG0P": "GPU Proximity 0 (internal EMC1414) cooked temp (DegC) (TG0P)",
	"TG0T": "GPU Die temp Trend (DegC) (TG0T)",
	"TG0d": "GPU Die raw temp (DegC) (TG0d)",
	"TG0p": "GPU Proximity 0 (internal EMC1414) raw temp (DegC) (TG0p)",
	"TG1D": "GPU B Die cooked temp (DegC) (TG1D)",
	"TG1F": "GPU digital die temp filtered (DegC) (TG1F)",
	"TG1H": "Heat Pipe Temp2(TG1H)",
	"TG1P": "GPU GDDR5 Proximity 1 cooked temp (DegC) (TG1P)",
	"TG1d": "GPU digital die temp Raw (without offset) (DegC) (TG1d)",
	"TG1p": "GPU GDDR5 Proximity 1 raw temp (DegC) (TG1p)",
	"TG2P": "GPU VR Proximity 2 cooked temp (DegC) (TG2P)",
	"TG2p": "GPU VR Proximity 2 raw temp (DegC) (TG2p)",
	"TG3P": "GPU FB VR Proximity 3 (via diode connected to EMC1428) cooked temp (DegC) (TG3P)",
	"TG3p": "GPU FB VR Proximity 3 (via diode connected to EMC1428) raw temp (DegC) (TG3p)",
	"TG4P": "GPU B VRAM Proximity cooked temp (DegC) (TG4P)",
	"TG5P": "GPU B VR Proximity cooked temp (DegC) (TG5P)",
	"TH0A": "Drive 0 OOBv3 absolute cooked temp A (DegC) (TH0A)",
	"TH0B": "Drive 0 OOBv3 absolute cooked temp B (DegC) (TH0B)",
	"TH0C": "Drive 0 OOBv3 absolute cooked temp C (DegC) (TH0C)",
	"TH0F": "Drive 0 OOBv3 relative cooked temp Max Filtered (DegC) (TH0F)",
	"TH0O": "Drive 0 OOBv1 mapped temp (DegC) (TH0O)",
	"TH0P": "Drive 0 OOBv1 filtered temp (DegC) (TH0P)",
	"TH0R": "SSD 0 (2.5) OOB v3 relative cooked temp Max (DegC) (TH0R)",
	"TH0V": "Drive 0 backup virtual DRAM temp (DegC) (TH0V)",
	"TH0X": "SSD 0 OOB v3 cooked temp Max (DegC) (TH0X)",
	"TH0a": "Drive0 (SSD Gumstick) OOB v3 raw temp A (DegC) (TH0a)",
	"TH0b": "Drive0 (SSD Gumstick) OOB v3 raw temp B (DegC) (TH0b)",
	"TH0c": "Drive0 (SSD Gumstick) OOB v3 raw temp C (DegC) (TH0c)",
	"TH0e": "Drive 0 temp sensor existence used for TM0V (0) = present (TH0e)",
	"TH0o": "Drive 0 (2.5) OOB v1 raw bit-mapped temp (TH0o)",
	"TH0x": "Drive0 (SSD Gumstick) OOB v3 raw temp Max (DegC) (TH0x)",
	"TH1A": "Drive 1 OOBv3 absolute cooked temp A (DegC) (TH1A)",
	"TH1B": "Drive 1 OOBv3 absolute cooked temp B (DegC) (TH1B)",
	"TH1C": "Drive 1 OOBv3 absolute cooked temp C (DegC) (TH1C)",
	"TH1F": "Drive 1 OOBv3 relative filtered temp Max (DegC) (TH1F)",
	"TH1O": "Drive 1 OOBv1 mapped temp (DegC) (TH1O)",
	"TH1P": "Drive 1 OOBv1 filtered temp (DegC) (TH1P)",
	"TH1R": "Drive 1 OOBv3 relative cooked temp Max (DegC) (TH1R)",
	"TH1X": "SSD 1 OOB v3 cooked temp Max (DegC) (TH1X)",
	"TH1a": "Drive 1 OOBv3 raw temp A (DegC) (TH1a)",
	"TH1b": "Drive 1 OOBv3 raw temp B (DegC) (TH1b)",
	"TH1c": "Drive 1 OOBv3 raw temp C (DegC) (TH1c)",
	"TH1o": "Drive 0 OOBv1 raw bit-mapped temp (TH1o)",
	"THFH": "Fin Stack Temp(THFH)",
	"THSP": "T29 Proxmity temp (DegC) (THSP)",
	"TI0P": "I/O Board Proximity cooked temp (DegC) (TI0P)",
	"TI0T": "Thunderbolt Proximity cooked temp (DegC) (TI0T)",
	"TI0p": "TBT Proximity raw temp (DegC) (TI0p)",
	"TI1P": "5V/3V3 VR Proximity cooked temp (DegC) (TI1P)",
	"TL0P": "LCD unfiltered front-of-screen cooked temp (DegC) (TL0P)",
	"TL0V": "LCD filtered front-of-screen hotspot virtual temp (DegC) (TL0V)",
	"TL0p": "LCD unfiltered front-of-screen raw temp (DegC) (TL0p)",
	"TL1P": "LCD TCON Proximity cooked temp (DegC) (TL1P)",
	"TL1V": "LCD TL1v-filtered center front-of-screen temp for color compensation (DegC) (TL1V)",
	"TL1p": "LCD TCON Proximity raw temp (DegC) (TL1p)",
	"TL1v": "LCD filtered center front-of-screen temp that feeds TL1V (DegC) (TL1v)",
	"TL2p": "LCD IR target hotspot calculated temp (DegC) (TL2p)",
	"TL3p": "LCD IR temp sensor die raw temp (DegC) (TL3p)",
	"TM0D": "DIMM 0 PECI cooked temp (DegC) (TM0D)",
	"TM0P": "DIMM 0/1 Top Proximity cooked temp (DegC) (TM0P)",
	"TM0S": "Synthetic DIMM Estimate(DegC) (TM0S)",
	"TM0V": "Synthetic DIMM Estimate(DegC) (TM0V)",
	"TM0p": "LPDDR proximity raw temp (DegC) (TM0p)",
	"TM1D": "DIMM 1 PECI cooked temp (DegC) (TM1D)",
	"TM1P": "DIMM 1 Proximity cooked temp (DegC) (TM1P)",
	"TM1a": "DIMM virtual temp for 2x DIMM Slot1+2+3+4 config (DegC) (TM1a)",
	"TM1p": "DIMM 1 Proximity raw temp (DegC) (TM1p)",
	"TM2D": "DIMM 2 PECI cooked temp (DegC) (TM2D)",
	"TM2P": "DIMM 2 Proximity cooked temp (DegC) (TM2P)",
	"TM2a": "DIMM virtual temp for 2x DIMM Slot1+3 config (DegC) (TM2a)",
	"TM2b": "DIMM virtual temp for 2x DIMM Slot1+2 or Slot2+3 config (DegC) (TM2b)",
	"TM2c": "DIMM virtual temp for 2x DIMM Slot1+4 or Slot2+4 config (DegC) (TM2c)",
	"TM2d": "DIMM virtual temp for 2x DIMM Slot3+4 config (DegC) (TM2d)",
	"TM2p": "DIMM 2 Proximity raw temp (DegC) (TM2p)",
	"TM3D": "DIMM 3 PECI cooked temp (DegC) (TM3D)",
	"TM3P": "DIMM 3 Proximity cooked temp (DegC) (TM3P)",
	"TM3a": "DIMM virtual temp for 3x DIMM Slot2+3+4 config (DegC) (TM3a)",
	"TM3b": "DIMM virtual temp for 3x DIMM Slot1+3+4 config (DegC) (TM3b)",
	"TM3c": "DIMM virtual temp for 3x DIMM Slot1+2+4 config (DegC) (TM3c)",
	"TM3d": "DIMM virtual temp for 3x DIMM Slot1+2+3 config (DegC) (TM3d)",
	"TM3p": "DIMM 3 Proximity raw temp (DegC) (TM3p)",
	"TM4a": "DIMM virtual temp for 4x DIMM Slot1+2+3+4 config (DegC) (TM4a)",
	"TMBS": "DIMM Bandwidth Estimate (DegC) (TMBS)",
	"TMCD": "Memory Controller/Integrated Graphics PECI Temp(TMCD)",
	"TMLB": "MLB Proximity temp(DegC) (TMLB)",
	"TMXD": "DIMM max cooked temp (DegC) (TMXD)",
	"TMXP": "DIMM Proximity max cooked temp (DegC) (TMXP)",
	"TN0D": "MCP89 Die temp (SIT! force bit 11) (DegC) (TN0D)",
	"TN0P": "MCP89 Die temp (SIT! force bit 12) (DegC) (TN0P)",
	"TN1D": "MCP internal die temp (SIU! force bit 1) (DegC) (TN1D)",
	"TP0P": "PCH Prox Temp(TP0P)",
	"TPCD": "PCH Die cooked temp (DegC) (TPCD)",
	"TPCd": "PCH digital die raw temp (DegC) (TPCd)",
	"TS0D": "Speaker Remote Temp (DegC) (TS0D)",
	"TS0P": "Speaker Proximity Temp (DegC) (TS0P)",
	"TS0V": "Synthetic bottom skin (DegC) (TS0V)",
	"TS0p": "Skin 0 raw temp (DegC) (TS0p)",
	"TS1D": "Speaker Remote 2 Temp (DegC) (TS1D)",
	"TS1P": "Speaker Proximity 2 Temp (DegC) (TS1P)",
	"TS2P": "S2 camera proximity temp (DegC) (TS2P)",
	"TTF0": "Fan Target Temp(TTF0)",
	"TV0P": "Vent proximity temp (TV0P)",
	"TVFP": "Filtered Vent proximity (TVFP)",
	"TW0P": "Wifi proximity temp (DegC) (TW0P)",
	"TW0p": "Wifi proximity raw temp (DegC) (TW0p)",
	"Ta0P": "air flow temp (DegC) (Ta0P)",
	"TaSP": "DCIn Air flow Adjusted Temperature (DegC) (TaSP)",
	"Tb0P": "Backlight Controller proximity cooked temp (DegC) (Tb0P)",
	"Tb0p": "Backlight Controller proximity raw temp (DegC) (Tb0p)",
	"Te0T": "PCIe Switch Diode cooked temp (DegC) (Te0T)",
	"Te0t": "TBT Thermal Diode raw temp (DegC) (Te0t)",
	"Th0N": "Nand temp (DegC) (Th0N)",
	"Th0P": "Heat Spreader proximity temp (Th0P)",
	"Th0R": "RIO temp (DegC) (Th0R)",
	"Th1H": "Right Fin stack temp (DegC) (Th1H)",
	"Th2H": "Left Fin Stack Temp(Th2H)",
	"Tm0P": "MLB Proximity 0 cooked temp: MLB top near power connector (DegC) (Tm0P)",
	"Tm0p": "MLB Proximity 0 raw temp (Tm0p)",
	"Tm1P": "MLB Proximity 1 cooked temp: MLB top near GPU (DegC) (Tm1P)",
	"Tm1p": "MLB Proximity 1 raw temp (Tm1p)",
	"Tm2P": "MLB Proximity 2 cooked temp: MLB bottom near CPU (DegC) (Tm2P)",
	"Tm2p": "MLB Proximity 2 raw temp (Tm2p)",
	"Tp0P": "Power supply proximity temp (Tp0P)",
	"Tp0T": "PSU Secondary H/S Diode cooked temp (DegC) (Tp0T)",
	"Tp1P": "Secondary power supply proximity (Tp1P)",
	"Tp2F": "Power Supply T2 Secondary Heatsink filtered temp (DegC) (Tp2F)",
	"Tp2H": "Power Supply T2 Secondary Heatsink cooked temp (DegC) (Tp2H)",
	"Tp2h": "Power Supply T2 Secondary Heatsink raw temp (DegC) (Tp2h)",
	"TpFP": "Filtered Power supply 1 proximity (TpFP)",
	"Ts0P": "Palm rest skin temp (SIT! force bit 13) (DegC) (Ts0P)",
	"Ts0S": "Synthetic, bottom case, skin temp (Ts0S)",
	"Ts1P": "Palm Rest Temp 2 (DegC) (Ts1P)",
	"Ts1S": "Synthetic top skin (DegC) (Ts1S)",
	"V50R": "5V lowside voltage (Volts) (V50R)",
	"VACC": "ACC voltage. (Volts) (VACC)",
	"VAPC": "Airport Voltage (DEBUG) (Volts) (VAPC)",
	"VBLC": "Backlight. (Volts) (VBLC)",
	"VC0C": "CPU Core voltage. (Volts) (VC0C)",
	"VC0G": "CPU AXG low-side voltage (Volts) (VC0G)",
	"VC0M": "CPU Mem low-side voltage (Volts) (VC0M)",
	"VC0S": "CPU VSA low side voltage (Volts) (VC0S)",
	"VC1C": "1.05 S0 voltage. (Volts) (VC1C)",
	"VC1R": "CPU High Side Voltage from EMC1704 (Volts) (VC1R)",
	"VC2C": "VCCSA voltage. (Volts) (VC2C)",
	"VCFR": "FIVR CPU supply voltage. (Volts) (VCFR)",
	"VCRC": "CPU Ripple voltage. (DEBUG) (Volts) (VCRC)",
	"VCRP": "CPU Ripple voltage. (DEBUG) (Volts) (VCRP)",
	"VCS0": "CPU Core voltage. (Volts) (VCS0)",
	"VD0R": "DC In voltage. (Volts) (VD0R)",
	"VD2R": "Power Supply 12V voltage (Volts) (VD2R)",
	"VG0C": "Ext GPU Core voltage. (Volts) (VG0C)",
	"VG0F": "GPU Frame Buffer input-side voltage (Volts) (VG0F)",
	"VG0I": "GPU VDDCI load-side voltage (Volts) (VG0I)",
	"VG0S": "GPU A VDDCI low side voltage (Volts) (VG0S)",
	"VG0U": "GPU Uncore high-side voltage (Volts) (VG0U)",
	"VG1C": "GPU Core input-side voltage (Volts) (VG1C)",
	"VG1F": "GPU Frame Buffer input-side voltage (Volts) (VG1F)",
	"VG1S": "GPU B VDDCI low side voltage (Volts) (VG1S)",
	"VH05": "HDD 5V low-side voltage (Volts) (VH05)",
	"VH0R": "SSD 3V3 low side voltage (Volts) (VH0R)",
	"VH1R": "SSD 3.3V low-side voltage (Volts) (VH1R)",
	"VL3t": "LCD IR temp sensor voltage (Raw Register) (VL3t)",
	"VM0C": "1.2V to CPU/MEM lowside voltage (Volts) . (VM0C)",
	"VM0R": "DIMM low-side voltage (Volts) (VM0R)",
	"VM1C": "1.8V S3 lowside voltage (Volts) . (VM1C)",
	"VN0C": "Int GPU core voltage. (Volts) (VN0C)",
	"VN0R": "AGX Core Voltage (Volts) (VN0R)",
	"VN1C": "MCP Memory Voltage (SPS! force bit 15) (Volts) (VN1C)",
	"VN1R": "PCH low-side voltage (Volts) (VN1R)",
	"VODC": "ODD Voltage (DEBUG) (Volts) (VODC)",
	"VP0R": "PBus Voltage (Volts) (VP0R)",
	"VP0T": "PSU cooked temperature represented as a voltage (Volts) (VP0T)",
	"VR1R": "PCH/GPU/TBT load-side voltage (Volts) (VR1R)",
	"VR3R": "3.3V S0 rail voltage (Volts) (VR3R)",
	"VSDC": "SD Card Voltage (Volts) (VSDC)",
	"VTPC": "T101 Boost/PBus Accuator Voltage. (Volts) (VTPC)",
	"VZAP": "Airport Voltage(VZAP)",
	"VZBL": "Backlight Voltage(VZBL)",
	"VZDM": "Memory Voltage (SPS! force bit 19) (Volts) (VZDM)",
	"VZHD": "SSD Voltage (SPS! force bit 18) (Volts) (VZHD)",
	"Vp0C": "ACDC cooked temp (Volts) (Vp0C)",

	// custom keys
	"Exhaust":      "Exhaust Fan Speed (RPM) (Exhaust)",
	"ThrottleTime": "Core throttle time in milliseconds",
}
