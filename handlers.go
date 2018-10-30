package main

import (
	"github.com/labstack/echo"
	"net/http"
)

func getTarget(c echo.Context) error {
	return c.String(http.StatusOK, string(target))
}

func postTarget(c echo.Context) error {
	return c.JSON(http.StatusAccepted, "POST Target")
}

func getSensors(c echo.Context) error {
	return c.JSON(http.StatusOK, "GET Sensors")
}

func postSensors(c echo.Context) error {
	return c.JSON(http.StatusAccepted, "POST sensors")
}
