# Requirements

## Backend

* Go
* MQTT broker ([mosquitto](https://hub.docker.com/_/eclipse-mosquitto) works well)
* npm

## MQTT broker setup

* Create a file called `mosquitto.conf` with the following contents:

```text
# Attention: setting to allow_anonymous to true will allow any client on your network to connect and read/write all MQTT messages
listener 1883
allow_anonymous true
```

* Run the below command in a separate window, replacing `MOSQUITTO_PATH` with the full (absolute) path to the `mosquitto.conf` file you just created:

```shell
docker run -it -p 1883:1883 -v MOSQUITTO_PATH:/mosquitto/config/mosquitto.conf eclipse-mosquitto
```

## Development

* Copy `garden.sample.ini` to `garden.ini` and set appropiate values for your environment
* Run `./build.sh && ./garden`
