package main

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
)

func msgHandler(_ mqtt.Client, m mqtt.Message) {
	logf(3, "MQTT message received")
	var (
		periph string
		tmp    float64
	)
	_, err := fmt.Sscan(string(m.Payload()), periph, tmp)
	if err != nil {
		logf(1, "Cannot parse MQTT message: %s", err)
		return
	}
	readings[periph] = tmp
}

func subscribeMQTT() {
	opts := mqtt.NewClientOptions().AddBroker("192.168.1.30") // Todo: Change to a static IP
	opts.SetClientID("thermoPi")

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		logf(-1, "Cannot connect to MQTT broker: %s", token.Error())
	}
	if token := c.Subscribe("temperature", 0, msgHandler); token.Wait() && token.Error() != nil {
		logf(-1, "Cannot subscribe to MQTT topic: %s", token.Error())
		panic(token.Error())
	}
}
