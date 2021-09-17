# Requirements

## Backend

* Go
* MQTT broker ([mosquitto](https://hub.docker.com/_/eclipse-mosquitto) works well)
* npm

## Development

* Set the following environmental variables:
  * `MQTT_HOST`: MQTT broker address
  * `MQTT_USERNAME`: MQTT username (if required)
  * `MQTT_PASSWORD`: MQTT password (if required)
* Run `go build && ./garden`
