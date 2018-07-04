package main

import (
	"github.com/stianeikeland/go-rpio"
	"fmt"
	"os"
	"time"
)

var (
	// Use mcu pin 10, corresponds to physical pin 19 on the pi
	pin = rpio.Pin(4)
)

func main() {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()

	// Toggle pin 20 times
	for x := 0; x < 20; x++ {
		pin.Toggle()
		time.Sleep(time.Second / 5)
	}
}
