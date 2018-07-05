#include <Arduino.h>
#include "Adafruit_BLE.h"
#include "Adafruit_BluefruitLE_UART.h"

#include "BluefruitConfig.h"

#define FACTORYRESET_ENABLE         1
#define MINIMUM_FIRMWARE_VERSION    "0.6.6"
#define MODE_LED_BEHAVIOUR          "MODE"

Adafruit_BluefruitLE_UART ble(BLUEFRUIT_HWSERIAL_NAME, BLUEFRUIT_UART_MODE_PIN);


int tempPin = A1;
int tempRead = 0;
float tempMvolts = 0;
float tempC = 0;

void error(const __FlashStringHelper*err) {
  Serial.println(err);
  while (1);
}

void setup() {
  analogReference(AR_DEFAULT);
  analogReadResolution(12);
  ble.begin(VERBOSE_MODE);
  if ( FACTORYRESET_ENABLE )
  {
    ble.factoryReset();
  }

  ble.echo(false);

//  ble.verbose(false);
  while (! ble.isConnected()) {
      delay(500);
  }

  if ( ble.isVersionAtLeast(MINIMUM_FIRMWARE_VERSION) )
  {
    ble.sendCommandCheckOK("AT+HWModeLED=" MODE_LED_BEHAVIOUR);
  }

  ble.setMode(BLUEFRUIT_MODE_DATA);
}

void loop() {
  tempRead = analogRead(tempPin);
  tempMvolts = tempRead * 3300 / 4096;
  tempC = (tempMvolts - 500) / 10;
  ble.println(String(tempC));
  delay(1000);
}
