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
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
)

func getTarget(c echo.Context) error {
	logf(3, "Responding to GET /target with %d", target)
	data := struct {
		Value int `json:"value"`
	}{target}
	return c.JSON(http.StatusOK, data)
}

func postTarget(c echo.Context) error {
	var err error
	r := c.Request()
	dec := json.NewDecoder(r.Body)
	t, err := dec.Token()
	if err != nil {
		return err
	}
	logf(3, "%T: %v\n", t, t)
	var m struct {
		Target    int  `json:"target"`
		Permanent bool `json:"permanent"`
	}
	for dec.More() {
		err := dec.Decode(&m)
		if err != nil {
			return err
		}
		logf(3, "Data: %v", m.Target)
	}
	target = m.Target
	if m.Permanent {
		setTarget(target)
	}
	logf(2, "Responding to POST /target from %s", r.Referer())
	return c.String(http.StatusAccepted, "Accepted")
}

func getSensors(c echo.Context) error {
	data := make([]struct {
		Name string  `json:"name"`
		Temp float64 `json:"value"`
	}, len(readings))
	i := 0
	for k, v := range readings {
		data[i].Name = k
		data[i].Temp = v
		i++
	}
	logf(3, "Responding to GET /sensors with: %v", data)
	return c.JSON(http.StatusOK, data)
}

func getSensorByID(c echo.Context) error {
	data := struct {
		Temp float64 `json:"value"`
	}{
		readings[c.Param("id")],
	}
	return c.JSON(http.StatusOK, data)
}

func postSensors(c echo.Context) error {
	return c.JSON(http.StatusAccepted, "POST sensors")
}
