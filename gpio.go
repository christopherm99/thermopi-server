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
