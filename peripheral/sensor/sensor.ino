#include <Wire.h>
#include <SPI.h>
#include <WiFiNINA.h>
#include <MQTT.h>
#include "Adafruit_MCP9808.h"

const char ssid[] = "SSID";
const char pass[] = "PASSWORD";
const char host[] = "192.168.1.30";
const char id[] = "ID";
int status = WL_IDLE_STATUS;

WiFiClient net;
MQTTClient client;
Adafruit_MCP9808 temp = Adafruit_MCP9808();

unsigned long lastMillis = 0;

void connect() {
    if (WiFi.status() == WL_NO_MODULE) {
        while(true);
    }
    while (status != WL_CONNECTED) {
        status = WiFi.begin(ssid, pass);
    }
}

void setup() {
    connect();
    temp.begin();
    client.begin(host, net);
    client.connect(id);
}

void loop() {
    client.loop();
    if (!client.connected()) {
        client.connect(id);
    }

    // publish a message roughly every second.
    if (millis() - lastMillis > 1000) {
        lastMillis = millis();
        client.publish("temperature", String(temp.readTempC()));
    }
}