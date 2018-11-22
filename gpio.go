package main

import (
	"encoding/csv"
	"github.com/stianeikeland/go-rpio"
	"os"
	"strconv"
	"time"
)

func setTarget(t int) {
	// Write to local (in memory) copy of schedule.
	schedule[time.Now().Weekday()][time.Now().Hour()] = t
	// Write to persistent (on drive) copy.
	// TODO: Add option to change schedule location
	file, err := os.OpenFile("schedule.csv", os.O_RDWR, 0644)
	if err != nil {
		logf(-1, "Cannot open schedule.csv: %s", err)
	}
	c := csv.NewWriter(file)
	data := make([][]string, len(schedule))
	for i := range data {
		data[i] = make([]string, len(schedule[i]))
	}
	for d := 0; d < len(schedule); d++ {
		for h := 0; d < len(schedule[d]); h++ {
			data[d][h] = strconv.Itoa(schedule[d][h])
		}
	}
	err = c.WriteAll(data)
	if err != nil {
		logf(0, "Cannot write new schedule: %s", err)
	}
}

func enable() {
	config.fanPin.Write(rpio.High)
	config.compPin.Write(rpio.High)
}

func disable() {
	config.fanPin.Write(rpio.Low)
	config.compPin.Write(rpio.Low)
}

func avgAmbient() float64 {
	avg := 0.0
	for _, v := range readings {
		avg += v
	}
	return avg / float64(len(readings))
}

func beginLogic() {
	// TODO: Add override option
	tick := time.NewTicker(config.lockout)
	for {
		<-tick.C
		if float64(target) < avgAmbient() {
			logf(2, "Enabling A/C...")
			enable()
		} else {
			logf(2, "Disabling A/C...")
			disable()
		}

	}
}
