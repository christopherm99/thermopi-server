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
	"encoding/csv"
	"time"
	"net/http"
	"html/template"
)


var (
	pin              = rpio.Pin(4)
	sensor1ServiceID = gatt.MustParseUUID("6e400001-b5a3-f393-e0a9-e50e24dcca9e")
	senor1TXCharID   = gatt.MustParseUUID("6e400003-b5a3-f393-e0a9-e50e24dcca9e")
	schedule         [7][24]int
	sensor1Temp      float64
	sensor2Temp      float64
	sensor3Temp      float64
)


func onStateChanged(d gatt.Device, s gatt.State) {
	//log.Println("State:", s)
	switch s {
	case gatt.StatePoweredOn:
		//log.Println("scanning...")
		d.Scan([]gatt.UUID{}, false)
		return
	default:
		d.StopScanning()
	}
}

func onPeriphDiscovered(p gatt.Peripheral, a *gatt.Advertisement, i int) {
	//fmt.Println("Discovered:", a.LocalName, ", with strength:", i)
	if p.ID() == "C9:6B:2C:72:BE:FA" {
		p.Device().StopScanning()
		p.Device().Connect(p)
	}
}

func onPeriphConnected(p gatt.Peripheral, err error) {
	//log.Printf("%v connected.\n", p.Name())

	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, service := range services {
		if service.UUID().Equal(sensor1ServiceID) {
			//log.Printf("Service Found %s\n", service.Name())

			characteristics, _ := p.DiscoverCharacteristics(nil, service)

			for _, characteristic := range characteristics {
				if characteristic.UUID().Equal(senor1TXCharID) {
					//log.Println("TX Characteristic Found")

					p.DiscoverDescriptors(nil, characteristic)

					p.SetNotifyValue(characteristic, onRecvMsg)
				}
			}
		}
	}
}

func onRecvMsg(c *gatt.Characteristic, b []byte, e error) {
	//fmt.Println("New message from: ", c.Name())
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
				//log.Printf("Got back %s", string(b))
				sensor1Temp = fltTemp
			}
		}
	}
}

func onPeriphDisconnected(p gatt.Peripheral, err error) {
	if err != nil {
		log.Println(err)
	}
	//log.Println("Disconnected")
	//log.Println("scanning...")
	p.Device().Scan([]gatt.UUID{}, false)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("public/index.html")
	if err != nil {
		panic(err)
	}
	data := struct {
		Goal int
		Time string
		Sensor1 string
		Sensor2 string
		Sensor3 string
		Sensor1Color string
		Sensor2Color string
		Sensor3Color string
	}{
		Goal:    schedule[time.Now().Weekday()][time.Now().Hour()],
		Time:    time.Now().Format(time.Kitchen),
	}
	data.Sensor1 = fmt.Sprint(int(sensor1Temp))
	data.Sensor2 = fmt.Sprint(int(sensor2Temp))
	data.Sensor3 = fmt.Sprint(int(sensor3Temp))
	fmt.Println(data.Sensor1)
	if int(sensor1Temp) > schedule[time.Now().Weekday()][time.Now().Hour()] {
		data.Sensor1Color = "#f44336"
	} else if int(sensor1Temp) == schedule[time.Now().Weekday()][time.Now().Hour()] {
		data.Sensor1Color = "#FF9800"
	} else {
		data.Sensor1Color = "#2196F3"
	}
	if int(sensor2Temp) > schedule[time.Now().Weekday()][time.Now().Hour()] {
		data.Sensor2Color = "#f44336"
	} else if int(sensor2Temp) == schedule[time.Now().Weekday()][time.Now().Hour()] {
		data.Sensor2Color = "#FF9800"
	} else {
		data.Sensor2Color = "#2196F3"
	}
	if int(sensor3Temp) > schedule[time.Now().Weekday()][time.Now().Hour()] {
		data.Sensor3Color = "#f44336"
	} else if int(sensor3Temp) == schedule[time.Now().Weekday()][time.Now().Hour()] {
		data.Sensor3Color = "#FF9800"
	} else {
		data.Sensor3Color = "#2196F3"
	}
	t.Execute(w, data)
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
			schedule[day][hour], err = strconv.Atoi(temp)
			if err != nil {
				panic(err)
			}
		}
	}

	// === BLUETOOTH LOW ENERGY CONNECTIONS ===
	var DefaultClientOptions = []gatt.Option{
		gatt.LnxMaxConnections(1),
		gatt.LnxDeviceID(-1, false),
	}
	//log.Println("Creating new Device...")
	d, err := gatt.NewDevice(DefaultClientOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return
	}
	//log.Println("Initializing handlers...")
	d.Handle(
		gatt.PeripheralDiscovered(onPeriphDiscovered),
		gatt.PeripheralConnected(onPeriphConnected),
		gatt.PeripheralDisconnected(onPeriphDisconnected),
	)
	//log.Println("Initializing Device...")
	d.Init(onStateChanged)
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", nil)
	for {
		if float64(schedule[time.Now().Weekday()][time.Now().Hour()]) > sensor1Temp- 5 {
			pin.High() // Switched because the relay doesn't work.
		} else if float64(schedule[time.Now().Weekday()][time.Now().Hour()]) > sensor1Temp+ 5 {
			pin.Low()
		}
		time.Sleep(time.Minute)
	}
}
