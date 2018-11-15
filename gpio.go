package main

import (
	"github.com/stianeikeland/go-rpio"
	"time"
)

func enable() {
	config.fanPin.Write(rpio.High)
	config.compPin.Write(rpio.High)
}

func disable() {
	config.fanPin.Write(rpio.Low)
	config.compPin.Write(rpio.Low)
}

func fan() {
	config.fanPin.Write(rpio.High)
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
	tick := time.NewTicker(config.lockout)
	for {
		<-tick.C
		if float64(target) < avgAmbient() {
			enable()
		} else {
			disable()
		}

	}
}
