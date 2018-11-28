package main

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
)

func mesgHandler(c mqtt.Client, m mqtt.Message) {
	fmt.Println(m)
}

func subscribeMQTT() {
	opts := mqtt.NewClientOptions().AddBroker("192.168.1.30") // Todo: Change to a static IP
	opts.SetClientID("thermoPi")

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	if token := c.Subscribe("temperature", 0, mesgHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
