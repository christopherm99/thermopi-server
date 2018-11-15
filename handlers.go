package main

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func getTarget(c echo.Context) error {
	logf(3, "Responding to GET /target with %d", target)
	data := struct {
		Value int `json:"value"`
	}{target}
	return c.JSON(http.StatusOK, data)
}

// TODO: Clean up this messy code. It is definitely inefficient and also very hard to read.
func postTarget(c echo.Context) error {
	r := c.Request()
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	logf(3, "Data POSTed to /target: %s", buf.String())
	var m struct {
		Target    int  `json:"target"`
		Permanent bool `json:"permanent"`
	}
	logf(3, "Now decoding POSTed data...")
	err := json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		logf(2, "Error parsing POST to /target: %s", err)
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
	logf(3, "Current sensor readings: %v", readings)
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
	val, err := strconv.ParseFloat(c.FormValue("value"), 64)
	if err != nil {
		logf(2, "Error parsing sensor data in /sensors: %s", err)
	}
	readings[c.Param("id")] = val
	return c.JSON(http.StatusAccepted, "POST sensors")
}
