package main

import (
	"github.com/paypal/gatt"
	"log"
	"github.com/stianeikeland/go-rpio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"regexp"
	"time"
	"encoding/csv"
)


type timeslot struct {
	begin time.Time
	end time.Time
	temp int
}


var (
	done = make(chan struct{})
	pin = rpio.Pin(4)
	uartServiceId = gatt.MustParseUUID("6e400001-b5a3-f393-e0a9-e50e24dcca9e")
	uartServiceTXCharId = gatt.MustParseUUID("6e400003-b5a3-f393-e0a9-e50e24dcca9e")
	schedule [7]timeslot
)


func onStateChanged(d gatt.Device, s gatt.State) {
	log.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		log.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, i int) {
	if p.ID() == "C9:6B:2C:72:BE:FA" {
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	log.Printf("%v connected.\n", p.Name())

	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, service := range services {
		if service.UUID().Equal(uartServiceId) {
			log.Printf("Service Found %s\n", service.Name())

			characteristics, _ := p.DiscoverCharacteristics(nil, service)

			for _, characteristic := range characteristics {
				if characteristic.UUID().Equal(uartServiceTXCharId) {
					log.Println("TX Characteristic Found")

					p.DiscoverDescriptors(nil, characteristic)

					p.SetNotifyValue(characteristic, onRecvMsg)
				}
			}
		}
	}
}

func onRecvMsg(c *gatt.Characteristic, b []byte, e error) {
	if e != nil {
		panic(e)
	}
	if len(b) > 4 {
		if len(b) == 7 {
			matched, err := regexp.Match("[[:digit:]]{2}.[[:digit:]]{2}\r\n", b)
			if err != nil {
				panic(err)
			}
			if matched {
				fltTemp, err := strconv.ParseFloat(strings.Trim(string(b), "\r\n"), 64)
				if err != nil {
					panic(err)
				}
				log.Printf("Got back %s", string(b))
				if fltTemp > 20 {
					pin.High()
				}
			}
		}
	}
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	log.Println("Disconnected")
	log.Println("scanning...")
	p.Device().Scan([]gatt.UUID{}, false)
}


func main() {
	// === RASPBERRY PI GPIO ===
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
	pin.Output()

	// === IMPORTING SCHEDULE ===
	file, err := os.Open("schedule.csv")
	if err != nil {
		panic(err)
	}
	csvRead := csv.NewReader(file)
	slcCsv, err := csvRead.ReadAll()
	if err != nil {
		panic(err)
	}
	for hour, row := range slcCsv[1:] {
		for day, temp := range row[1:] {
			schedule[day].begin = time.
		}
	}

	// === BLUETOOTH LOW ENERGY CONNECTIONS ===
	var DefaultClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, false),
	}
	log.Println("Creating new Device...")
	d, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	log.Println("Initializing handlers...")
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)
	log.Println("Initializing Device...")
	d.Init(onStateChanged)
	<-done
	log.Println("Done")
}
