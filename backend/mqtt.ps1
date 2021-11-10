docker run -it -p 1883:1883 -v "$(Get-Item .\mosquitto.conf | % { $_.FullName }):/mosquitto/config/mosquitto.conf" eclipse-mosquitto
