package main

import (
	"encoding/csv"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
	"github.com/stianeikeland/go-rpio"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	pin     = rpio.Pin(4)
	logFile = "/var/log/thermoPi/thermoPi.log"
)

var (
	target   int
	schedule [7][24]int
	//readings map[string]float64
	config = struct {
		lockout   time.Duration
		compPin   rpio.Pin
		fanPin    rpio.Pin
		sensorIDs []string
		verbosity int
	}{
		time.Minute,
		rpio.Pin(0),
		rpio.Pin(0),
		[]string{},
		0,
	}
)

func initConfig() {
	configFolder := os.Getenv("XDG_CONFIG_HOME")
	if configFolder == "" {
		logf(2, "$XDG_CONFIG_HOME unset. Using $HOME/.config as config root.")
		configFolder = path.Join(os.Getenv("HOME"), ".config")
	}
	err := os.MkdirAll(path.Join(configFolder, "thermoPi"), os.ModeDir)
	if err != nil {
		logf(-1, "Unable to create directory: %s", err)
	}
	if _, err := os.Stat(path.Join(configFolder, "/thermoPi/thermoPi.conf")); os.IsNotExist(err) {
		f, err := os.OpenFile(path.Join(configFolder, "/thermoPi/thermoPi.conf"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logf(-1, "Can't create thermoPi.conf: %s", err)
		}
		defer f.Close()
		if _, err := f.Write([]byte(`[thermoPi]
lockout   = "10m"
# compPin   = 0 # Set these values to the correct BCM pins.
# fanPin    = 0
# sensorIDs = [ "Kitchen", "Bedroom" ]
verbosity = 1`)); err != nil {
			logf(-1, "Can't write to thermoPi.conf: %s", err)
		}
		logf(-1, "Edit %s such that it reflects the proper values.", path.Join(configFolder, "/thermoPi/thermoPi.conf"))
	}
	var data map[string]struct {
		Lockout   string   `toml:"lockout"`
		CompPin   int      `toml:"compPin"`
		FanPin    int      `toml:"fanPin"`
		SensorIDs []string `toml:"sensorIDs"`
		Verbosity int      `toml:"verbosity,omitempty"`
	}
	_, err = toml.DecodeFile(path.Join(configFolder, "/thermoPi/thermoPi.conf"), &data)
	if err != nil {
		logf(-1, "Error reading thermoPi.conf: %s", err)
	}
	config.lockout, err = time.ParseDuration(data["thermoPi"].Lockout)
	if err != nil {
		logf(-1, "Unable to read lockout value from thermoPi.conf: %s", err)
	}
	logf(3, "The timeout is set to %s", config.lockout)
	config.compPin = rpio.Pin(data["thermoPi"].CompPin)
	logf(3, "The compressor pin is set to %s", config.compPin)
	config.fanPin = rpio.Pin(data["thermoPi"].FanPin)
	logf(3, "The fan pin is set to %s", config.fanPin)
	config.sensorIDs = data["thermoPi"].SensorIDs
	logf(3, "The list of sensor IDs is: %v", config.sensorIDs)
	config.verbosity = data["thermoPi"].Verbosity
}

func initSchedule() {
	file, err := os.Open("schedule.csv")
	if err != nil {
		panic(err)
	}
	csvRead := csv.NewReader(file)
	slcCsv, err := csvRead.ReadAll()
	if err != nil {
		panic(err)
	}
	for hour, row := range slcCsv[1:] {
		for day, temp := range row[1:] {
			schedule[day][hour], err = strconv.Atoi(temp)
			if err != nil {
				panic(err)
			}
		}
	}

	target = schedule[time.Now().Weekday()][time.Now().Hour()]

	tick := time.NewTicker(time.Hour)
	go func() {
		for {
			t := <-tick.C
			target = schedule[t.Weekday()][t.Hour()]
		}
	}()
}

func initGPIO() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pin.Output()
}

func initEcho() {
	e := echo.New()
	/* API:
	 * NOTE: All temperatures will be in centigrade.
	 *
	 * /target      GET  - Get current target temperature for this time slot.
	 *   Response Format Example:
	 *     { "target":25 }
	 *
	 * /target      POST - Set current target temperature for this time slot.
	 *   Request Format: A POST request with the following parameters:
	 *     target    - The new target temperature
	 *     permanent - Whether this change ought to be updated in the permanent schedule.
	 *                 Defaults to off.
	 *   Possible Responses:
	 *     202 - The POST request was accepted and will be reflected soon.
	 *     400 - The POST request is malformed (eg. too large a value) and will not be reflected.
	 *     5xx - The POST request was valid, but the server had an error.
	 *
	 * /sensors     GET  - Get list of active sensors' ids.
	 *   Response Format Example:
	 *     [
	 *       "Bedroom",
	 *       "Kitchen",
	 *       "Living Room"
	 *     ]
	 *
	 * /sensors/:id GET  - Get most recent temperature reading from :id sensor.
	 *   Response Format Example:
	 *     { "value":22 }
	 *
	 * /sensors/:id POST - Receive data from :id sensor. (NB: This will probably be replaced with MQTT in the future).
	 *   Request Format: A POST request with the following parameters:
	 *     value - The most recent temperature reading from the sensor.
	 *
	 */
	e.GET("/target", getTarget)
	e.POST("/target", postTarget)
	e.GET("/sensors", getSensors)
	e.POST("/sensors", postSensors)
}

func main() {
	if err := os.Remove(logFile); err != nil {
		log.Fatalln(err)
	}
	logf(1, "Message Key: (EE) - Error, (WW) - Warning, (II) - Information, (DD) - Debug, (VV) - Verbose")
	initConfig()
	initSchedule()     // Read schedule from CSV and start target loop.
	initGPIO()         // Setup Raspberry Pi's GPIO pins for access and begin thermostat logic.
	defer rpio.Close() // Remember to close Raspberry Pi's GPIO pins when done.
	initEcho()         // Setup Echo web server.
}
