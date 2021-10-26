# ESP8266 Firmware

This is the firmware for creating a garden sensor system. These instructions have not been tested on Windows, but pull requests for improving them are welcome.

## Requirements

* ESP8266 development board
* PlatformIO

## Developing

### First steps

1. Install PlatformIO. This can be done in one of two ways:
    1. Install the extension in VSCode (easiest).

    2. Install the PlatformIO Core CLI system wide using [these](https://docs.platformio.org/en/latest/core/installation.html) instructions.

2. Generate the JSON configuration needed to configure a system. This configuration can be created manually or through the web interface.

    1. To use the web interface, click the green floating plus button in the lower right hand corner, fill out the configuration editor form, and click the save button. The JSON configuration will appear at the bottom of the form.

    2. Alternatively, you can edit the below JSON block with a text editor and modify the empty strings so they fit your environment.

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

1. Open a PlatformIO CLI terminal and run `pio run -e esp8266_debug -t upload`

2. Once the firmware finishes compiling and uploading, run `pio device monitor`

3. Paste in the JSON block you created earlier.

4. Restart the ESP chip either by entering `{"Command":"restart"}` or by pressing the reset button on the board.
