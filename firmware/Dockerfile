FROM debian:bullseye-slim

RUN apt-get update && apt-get upgrade -y
RUN apt-get install -y curl unzip python3

# ===== esptool =====
RUN curl -L https://github.com/espressif/esptool/releases/download/v3.1/esptool-v3.1-linux-amd64.zip -o /tmp/esptool.zip

# Extract all esptool scripts into /usr/local/sbin/, discarding the paths embedded in the archive.
RUN unzip -j /tmp/esptool.zip -d /usr/local/sbin/ && chmod +x /usr/local/sbin/esptool

# ===== Arduino CLI ======
RUN curl -L https://github.com/arduino/arduino-cli/releases/download/0.19.0/arduino-cli_0.19.0_Linux_64bit.tar.gz -o /tmp/arduino.tar.gz
RUN cd /usr/local/sbin && tar -xzvf /tmp/arduino.tar.gz && mv arduino-cli arduino && chmod +x arduino && cd -
RUN arduino core install esp8266:esp8266 --additional-urls https://arduino.esp8266.com/stable/package_esp8266com_index.json

# ===== PubSubClient library =====
RUN curl -L https://github.com/Imroy/pubsubclient/archive/refs/heads/master.zip -o /tmp/pubsub.zip
WORKDIR /root/Arduino/libraries
RUN unzip /tmp/pubsub.zip && mv pubsubclient-master PubSubClient

# ===== Volume and defaults setup =====
ENV BAUD=460800

# /tmp is a volume as the arduino CLI caches intermediate build files here.
# Mounting a volume here speeds up repeated builds 3x.
VOLUME ["/firmware", "/tmp"]

COPY docker/* /usr/local/sbin/

CMD build
