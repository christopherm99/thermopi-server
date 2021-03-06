# ThermoPi 
[![Travis (.org)](https://img.shields.io/travis/christopherm99/thermopi-server.svg?style=for-the-badge)](https://travis-ci.org/christopherm99/thermopi-server)
[![GitHub](https://img.shields.io/github/license/christopherm99/thermopi-server.svg?style=for-the-badge)](https://github.com/christopherm99/thermopi-server/blob/master/LICENSE)
[![Coveralls github](https://img.shields.io/coveralls/github/christopherm99/thermopi-server.svg?style=for-the-badge)](https://coveralls.io/github/christopherm99/thermopi-server)
 * [Description](#description)
 * [Hardware](#hardware)
 * [Installation](#installation)
    * [Arduino Code](#arduino-code)
 * [Configuration](#configuration)
    * [Lockout Setting](#lockout-setting)
    * [Pin Settings](#pin-settings)
    * [Verbosity Settings](#verbosity-settings)
 * [Running](#running)
 * [Development](#development)
    * [Installation for Development](#contributing)
    * [API Specification](#api-specification)
        * [Frontend Development](#for-frontend-development)
        * [Backend Development](#for-sensor-development)
    * [Server Modularity](#server-modularity)
## Description
ThermoPi is attempt to make a simple, extensible, and powerful thermostat framework for the Raspberry Pi. The backend is
written in Golang, using Echo to power the server, and Stianeikeland's go-rpio library to control the Raspberry Pi's
GPIO pins. The frontend is written using Vue.js and TypeScript, and can be found 
[here](https://github.com/christopherm99/thermopi-webapp). Finally, the code for the peripheral sensors is written in
Arduino C++, using an Adafruit library to interface with their thermometer module, the MCP9808.
## Hardware
ThermoPi has been developed for the Raspberry Pi 3B+, with the official Raspberry Pi touchscreen, which can be found
[here](https://www.raspberrypi.org/products/raspberry-pi-3-model-b-plus/) and 
[here](https://www.raspberrypi.org/products/raspberry-pi-touch-display/), respectively, to host the Golang server. The
code for the peripheral sensors has been designed to run on the Arduino MKR 1010, and the Adafruit MCP9890, found 
[here](https://store.arduino.cc/usa/arduino-mkr-wifi-1010) and [here](https://www.adafruit.com/product/1782).
Both of these systems can likely be switched out with alternatives, provided there is some tweaking of the respective
code.
## Installation
If you are looking to find out how to contribute to thermoPi, see [here](#contributing).

Ensure you have at least [Golang 1.11](https://golang.org/) and [Dep](https://github.com/golang/dep) installed. For the 
default frontend, also install at least [Node v11](https://nodejs.org).
Then run this on the server:
```bash
git clone https://github.com/christopherm99/thermopi-server.git thermopi
cd thermopi
[sudo] ./install.sh
```
The install script should tell you where thermoPi has been installed to. 

Note: It is possible that thermoPi will work with different versions of Go and NodeJS, but they have not been tested.
### Arduino Code
To flash the arduino code to your device, follow these steps
1. Install the [Arduino IDE](https://www.arduino.cc/en/Main/Software)
2. Install the package for the MKR 1010 in the IDE, via Tools > Board > Boards Manager, and searching for Arduino SAMD.
3. Install the Adafruit_MCP9808.h library, via Tools > Manage Libraries, and searching for Adafruit MCP9808.
4. Select the newly installed board from Tools > Board > Arduino MKR WiFi 1010.
5. Select the correct port for the board (plug in the cable to the Arduino first) from Tools > Port.
6. Open the code for the peripheral from Files > Open, and navigating to the peripheral subdirectory of the GitHub repo.
7. Edit lines 6 and 7 so that they reflect the proper values.
8. Flash the code by clicking the upload button (➡) at the top left corner of the editor. 
## Configuration
ThermoPi can be configured using either commandline flags, as defined in usage or with a configuration file. ThermoPi 
will check for a config file at `$XDG_CONFIG_HOME/thermoPi/thermoPi.conf`, or if `$XDG_CONFIG_HOME` is unset, at 
`$HOME/.config/thermoPi/thermoPi.conf`. If not found, thermoPi will create a template file there. The config file 
should follow this format. See below for variable explanation.
```toml
[thermoPi]
lockout   = "10m10s"
compPin   = 4 
fanPin    = 5 
verbosity = 1 
```
### Lockout Setting
To prevent burnout of the air compressor, thermoPi has a lockout time period, which defaults to 1 minute. This means
if the signal is given to the compressor to start cooling, thermoPi must wait this amount of time before turning off
the compressor, and vise versa.
### Pin Settings
For the compressor and fan pin values, see [this](https://pinout.xyz/) diagram to find out which BCM pins you are using.
### Verbosity Settings
* Level -1 (Silent): Only fatal errors are displayed.
* Level 0 (Quiet): Only errors are displayed. 
* Level 1 (Normal): Only errors and normal messages displayed (eg. the server port and ip).
* Level 2 (Debug): Only errors, messages, and warnings displayed.
* Level 3 (Verbose): All possible messages displayed.
## Running
To run the server, first install it, as explained in the Installation section. The output of install.sh should give you
a file to run (/usr/local/bin in most cases). Run this to start the server.
## Development
### Contributing
For contributors, the process of setting up the repository is more complicated. However, one must still ensure they have
at least [Golang 1.11](https://golang.org/) and [Dep](https://github.com/golang/dep) installed. Then run the following 
from your ${GOPATH}/src directory:
```bash
git clone https://github.com/christopherm99/thermopi-server.git
dep ensure
```
The repository will now be set up for development.
### API Specification
See https://thermopi.docs.apiary.io
### Server Modularity:
The Golang server does not require the Vue.js webapp to be served, and anything could be served by the Echo server, in
theory. To serve different files, select no when prompted by `install.sh` to install the default webapp and place your 
`index.html` in `/usr/share/thermoPi/dist/`.
