package main

import (
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
	t, err = dec.Token()
	if err != nil {
		return err
	}
	logf(3, "%T: %v\n", t, t)
	logf(2, "Responding to POST /target from %", r.Referer())
	target, err = strconv.Atoi(c.FormValue("target"))
	if err != nil {
		logf(0, "Error parsing new target data: %s", err)
	}
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
