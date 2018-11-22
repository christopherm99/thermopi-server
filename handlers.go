package main

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func getTarget(c echo.Context) error {
	logf(3, "GET /target received")
	data := struct {
		Value int `json:"value"`
	}{target}
	return c.JSON(http.StatusOK, data)
}

// TODO: Clean up this messy code. It is definitely inefficient and also very hard to read.
func postTarget(c echo.Context) error {
	logf(3, "POST /target received")
	r := c.Request()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logf(1, "Error parsing POST /target: %s", err)
	}
	var m struct {
		Target    int  `json:"target"`
		Permanent bool `json:"permanent"`
	}
	logf(3, "Now decoding POSTed data...")
	err = json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		logf(1, "Error parsing POST /target: %s", err)
	}
	target = m.Target
	if m.Permanent {
		setTarget(target)
	}
	return c.String(http.StatusAccepted, "Accepted")
}

func getSensors(c echo.Context) error {
	logf(3, "GET /sensors received")
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
	return c.JSON(http.StatusOK, data)
}

func getSensorByID(c echo.Context) error {
	logf(3, "GET /sensors/:id received")
	data := struct {
		Temp float64 `json:"value"`
	}{
		readings[c.Param("id")],
	}
	return c.JSON(http.StatusOK, data)
}

func postSensors(c echo.Context) error {
	logf(3, "POST /sensors received")
	val, err := strconv.ParseFloat(c.FormValue("value"), 64)
	if err != nil {
		logf(2, "Error parsing POST /sensors: %s", err)
	}
	readings[c.Param("id")] = val
	return c.JSON(http.StatusAccepted, "POST sensors")
}
