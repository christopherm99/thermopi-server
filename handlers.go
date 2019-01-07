package main

import (
	"encoding/json"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func getTarget(c echo.Context) error {
	data := struct {
		Value int `json:"value"`
	}{target}
	return c.JSON(http.StatusOK, data)
}

func postTarget(c echo.Context) error {
	dec := json.NewDecoder(c.Request().Body)
	var m struct {
		Value     int  `json:"value"`
		Permanent bool `json:"permanent"`
	}
	if err := dec.Decode(&m); err != nil {
		logf(1, "Cannot parse POST /target: %s", err)
		return c.NoContent(http.StatusBadRequest)
	}
	target = m.Value
	logf(3, "POST /target received: %v", m.Value)
	if m.Permanent {
		setTarget(target)
	}
	return c.NoContent(http.StatusAccepted)
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
		logf(1, "Cannot parse POST /sensors: %s", err)
		return c.NoContent(http.StatusBadRequest)
	}
	readings[c.Param("id")] = val
	return c.NoContent(http.StatusAccepted)
}
