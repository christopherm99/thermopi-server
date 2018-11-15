package main

import (
	"encoding/csv"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/stianeikeland/go-rpio"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

const (
	logFile = "/var/log/thermoPi/thermoPi.log"
)

var (
	target   int                        // The target temperature for the thermostat to reach
	schedule [7][24]int                 // The parsed schedule from the CSV
	readings = make(map[string]float64) // The most recent readings from the sensors (eg. readings["Bedroom"] = 26)
	config   = struct {
		lockout   time.Duration
		compPin   rpio.Pin
		fanPin    rpio.Pin
		verbosity int
	}{
		time.Minute,
		rpio.Pin(0),
		rpio.Pin(0),
		3,
	}
)

func initConfig() {
	configFolder := os.Getenv("XDG_CONFIG_HOME")
	if configFolder == "" {
		logf(2, "$XDG_CONFIG_HOME unset. Using %s/.config as config root.", os.Getenv("HOME"))
		configFolder = path.Join(os.Getenv("HOME"), ".config")
	}
	err := os.MkdirAll(path.Join(configFolder, "thermoPi"), 0770)
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
	logf(3, "The compressor pin is set to %d", config.compPin)
	config.fanPin = rpio.Pin(data["thermoPi"].FanPin)
	logf(3, "The fan pin is set to %d", config.fanPin)
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

func initEcho() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
	// API
	e.GET("/target", getTarget)
	e.POST("/target", postTarget)
	e.GET("/sensors", getSensors)
	e.GET("/sensors/:id", getSensorByID)
	e.POST("/sensors", postSensors)
	// Webapp
	e.File("/", "/usr/share/thermoPi/dist/index.html")
	e.Static("/", "/usr/share/thermoPi/dist/")
	logf(-1, "Error within Echo: %s", e.Start(":8080"))
}

func main() {
	// DEBUG STUFF
	readings["Bedroom"] = 20
	readings["Kitchen"] = 30

	if _, err := os.Stat(logFile); !os.IsNotExist(err) {
		if err := os.Remove(logFile); err != nil {
			fmt.Println("Printing on line 146")
			log.Fatalln(err)
		}
	}
	fmt.Println("Message Key: (EE) - Error, (WW) - Warning, (II) - Information, (DD) - Debug, (VV) - Verbose")
	time.Sleep(100 * time.Millisecond)
	initConfig()
	initSchedule() // Read schedule from CSV and start target loop.
	//initGPIO()         // Setup Raspberry Pi's GPIO pins for access and begin thermostat logic.
	defer rpio.Close() // Remember to close Raspberry Pi's GPIO pins when done.
	time.Sleep(time.Millisecond)
	initEcho() // Setup Echo web server.
	beginLogic()
}
