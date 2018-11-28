package main

import (
	"encoding/csv"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/stianeikeland/go-rpio"
	"os"
	"path"
	"strconv"
	"time"
)

var (
	logFile      string
	scheduleFile string
	target       int                        // The target temperature for the thermostat to reach
	schedule     [7][24]int                 // The parsed schedule from the CSV
	readings     = make(map[string]float64) // The most recent readings from the sensors (eg. readings["Bedroom"] = 26)
	config       struct {
		lockout   time.Duration
		compPin   rpio.Pin
		fanPin    rpio.Pin
		verbosity int
		cors      bool
	}
	defaultConfig = []byte(`
[thermoPi]
lockout   = "10m" # Set this to amount of time to lockout between turning the A/C on and off.
compPin   = 0     # Set this and the following values to the correct BCM pins.
fanPin    = 0
verbosity = 1	  # Set this to the desired verbosity level (see README.md).
schedule  = "~/.config/thermoPi/" # Set this to your schedule.csv location.
CORS	  = false # Set this true to enable CORS (if you are hosting an external website).
keepLogs  = false # Set this to tell thermoPi to save old logs.
`)
)

func initConfig() {
	configFolder := os.Getenv("XDG_CONFIG_HOME")
	if configFolder == "" {
		logf(2, "$XDG_CONFIG_HOME unset. Using %s/.config as config root.", os.Getenv("HOME"))
		configFolder = path.Join(os.Getenv("HOME"), ".config")
	}
	// Create configuration files if nonexistent
	err := os.MkdirAll(path.Join(configFolder, "thermoPi"), 0770)
	if err != nil {
		logf(-1, "Cannot create configuration directory: %s", err)
	}
	if _, err := os.Stat(path.Join(configFolder, "/thermoPi/thermoPi.toml")); os.IsNotExist(err) {
		f, err := os.OpenFile(path.Join(configFolder, "/thermoPi/thermoPi.toml"), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logf(-1, "Cannot create thermoPi.toml: %s", err)
		}
		defer func() {
			err := f.Close()
			if err != nil {
				logf(-1, "Cannot close thermoPi.toml: %s", err)
			}
		}()
		if _, err := f.Write(defaultConfig); err != nil {
			logf(-1, "Cannot write to thermoPi.toml: %s", err)
		}
		logf(-1, "Exiting... Edit %s such that it reflects the proper values.", path.Join(configFolder, "/thermoPi/thermoPi.toml"))
	} else if err != nil {
		logf(-1, "Cannot stat thermoPi.toml: %s", err)
	}
	// Temporary data holder
	var data map[string]struct {
		Lockout   string `toml:"lockout"`
		CompPin   int    `toml:"compPin"`
		FanPin    int    `toml:"fanPin"`
		Verbosity int    `toml:"verbosity"`
		Schedule  string `toml:"schedule"`
		CORS      bool   `toml:"CORS"`
		Logs      bool   `toml:"persistentLogs"`
	}
	// Decoding configuration
	_, err = toml.DecodeFile(path.Join(configFolder, "/thermoPi/thermoPi.toml"), &data)
	if err != nil {
		logf(-1, "Cannot read thermoPi.toml: %s", err)
	}
	// Setup log files.
	if data["thermoPi"].Logs {
		logFile = fmt.Sprintf("%s-%d.log", path.Join(os.Getenv("HOME"), "/.cache/thermoPi/thermoPi"), time.Now().Unix())
	} else {
		logFile = path.Join(os.Getenv("HOME"), "/.cache/thermoPi/thermoPi.log")
		if _, err := os.Stat(logFile); !os.IsNotExist(err) {
			if err := os.Remove(logFile); err != nil {
				logf(-1, "Cannot delete old logs: %s", err)
			}
		}
	}
	// Setup config values
	config.lockout, err = time.ParseDuration(data["thermoPi"].Lockout)
	if err != nil {
		logf(-1, "Cannot read lockout value from thermoPi.toml: %s", err)
	}
	logf(3, "The timeout is set to %s", config.lockout)
	config.compPin = rpio.Pin(data["thermoPi"].CompPin)
	logf(3, "The compressor pin is set to %d", config.compPin)
	config.fanPin = rpio.Pin(data["thermoPi"].FanPin)
	logf(3, "The fan pin is set to %d", config.fanPin)
	config.cors = data["thermoPi"].CORS
	if config.cors {
		logf(2, "CORS is enabled")
	}
	config.verbosity = data["thermoPi"].Verbosity
	// Setup schedule location
	scheduleFile = data["thermoPi"].Schedule
	logf(3, "The schedule file is set to %s", scheduleFile)
}

func initSchedule() {
	file, err := os.Open(scheduleFile)
	if err != nil {
		logf(-1, "Cannot open schedule.csv: %s", err)
	}
	csvRead := csv.NewReader(file)
	slcCsv, err := csvRead.ReadAll()
	if err != nil {
		logf(-1, "Cannot read schedule.csv: %s", err)
	}
	for hour, row := range slcCsv[1:] {
		for day, temp := range row[1:] {
			schedule[day][hour], err = strconv.Atoi(temp)
			if err != nil {
				logf(-1, "Cannot parse schedule.csv: %s", err)
			}
		}
	}

	target = schedule[time.Now().Weekday()][time.Now().Hour()]

	tick := time.NewTicker(time.Hour)
	go func() {
		logf(2, "Schedule ticker beginning")
		for {
			t := <-tick.C
			target = schedule[t.Weekday()][t.Hour()]
		}
	}()
}

func initEcho() {
	e := echo.New()
	if config.cors {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		}))
	}
	// API
	e.GET("/target", getTarget)
	e.POST("/target", postTarget)
	e.GET("/sensors", getSensors)
	e.GET("/sensors/:id", getSensorByID)
	e.POST("/sensors/:id", postSensors)
	// Webapp
	e.File("/", "/usr/share/thermoPi/dist/index.html")
	e.Static("/", "/usr/share/thermoPi/dist/")
	e.HideBanner = true
	e.Debug = config.verbosity > 1 // Enable Echo's debug mode if verbosity is higher than information level.
	logf(-1, "Cannot serve data: %s", e.Start(":8080"))
}

func main() {
	// TODO: Remove this once in prod.
	readings["Bedroom"] = 20
	readings["Kitchen"] = 30

	fmt.Println("Message Key: (EE) - Error, (WW) - Warning, (II) - Information, (DD) - Debug, (VV) - Verbose")
	time.Sleep(100 * time.Millisecond)
	initConfig()
	initSchedule() // Read schedule from CSV and start target loop.
	defer func() {
		err := rpio.Close() // Remember to close Raspberry Pi's GPIO pins when done.
		if err != nil {
			logf(-1, "Error closing GPIO outputs: %s", err)
		}
	}()
	time.Sleep(time.Millisecond)
	initEcho() // Setup Echo web server.
	beginLogic()
}
