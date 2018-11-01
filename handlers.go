package main

import (
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func getTarget(c echo.Context) error {
	logf(3, "Responding to GET /TARGET with %d", target)
	data := struct {
		Value int `json:"value"`
	}{target}
	return c.JSON(http.StatusOK, data)
}

func postTarget(c echo.Context) error {
	var err error
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

func postSensors(c echo.Context) error {
	return c.JSON(http.StatusAccepted, "POST sensors")
}
