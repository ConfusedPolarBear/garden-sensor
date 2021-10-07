# ESP8266 Firmware

This is the firmware for creating a garden sensor system. These instructions have not been tested on Windows, but pull requests for improving them are welcome.

## Requirements

* ESP8266 development board
* [Arduino CLI](https://github.com/arduino/arduino-cli/releases)
  * Must be in the system path as `arduino`
* [esptool](https://github.com/espressif/esptool/releases), either installed from the repository, pip, or your system's package manager
  * Must be in the system path as `esptool.py`

## Initial setup

* Install the Arduino CLI
* Add the ESP8266 arduino core library from [here](https://arduino-esp8266.readthedocs.io/en/latest/installing.html).
* Add the PubSubClient library from [here](https://github.com/Imroy/pubsubclient).

## Developing

### First steps

1. Install Docker

2. Add your user to the docker and dialout group with `sudo usermod -aG docker USERNAME && sudo usermod -aG dialout USERNAME`.
    1. The `docker` and `dialout` groups grant your user permissions to access the Docker daemon and serial devices respectively.

3. While inside the `firmware` directory, run `docker build . -t garden-firmware`

4. Insert the below JSON block into a text editor and modify the empty strings so they fit your environment. You do not need to save this block anywhere.

```json
{
    "WifiSSID": "",
    "WifiPassword": "",
    "MQTTHost": "",
    "MQTTUsername": "",
    "MQTTPassword": ""
}
```

### Development

1. Run `./build.sh` to compile the firmware. Compiled binaries are in `build/esp8266.esp8266.generic`.

2. Run `./flash.sh PATH_TO_SERIAL_ADAPTER` to flash the newly compiled firmware onto a connected ESP8266 chip.

3. Open the serial port is configured at 115,200 bps 8N1. You should see some text followed by the prompt `Setup:`

4. Paste in the JSON block and press Enter. The system should respond with text similar to the following:

```json
{"Success":true,"Message":"","ConfiguredWiFi":true,"ConfiguredMQTT":true,"MQTTAuthenticated":true}
```

5. Restart the ESP chip either by entering `{"Command":"restart"}` or by pressing the reset button on the board.
