# ESP32 binaries

## Description
This document is intended for developers. If you are a user trying to flash your ESP32, download the ZIP archive and extract it.

## Contents of esp32.zip
The `esp32.zip` archive contains the three binary files that are needed for the ESP32 to boot:

| Filename               | Description                                                                 |
| ---------------------- | --------------------------------------------------------------------------- |
| boot_app0.bin          | Instructs the ESP32 to boot from the first app partition.                   |
| bootloader_dio_40m.bin | Second stage bootloader which enables flash partitioning, OTA updates, etc. |
| partitions.bin         | Partition table. See below for more details.                                |

## Using a custom archive
As with all other components of this project, it is possible to use a custom archive of ESP32 binary images instead of the official one. The following steps were used to create the official archive:

1. Compile a bootloader for your ESP32 chip. Instructions on how to do so are outside the scope of this document. This will be flashed to flash address `0x1000`.

2. Create a partition table and convert it into the binary format accepted by the bootloader. This will be flashed to flash address `0x8000`.

3. Use `boot_app0.bin` to boot from the `app0` partition. This will be flashed to flash address `0xe000`.

4. If your filenames differ from the ones in the default manifest, update the `manifest.json` file.

If you only want to use these files on your own server, you can place them in `data/firmware/esp32` to use them with the web based firmware install tool. However, if you want to make the files accessible by other garden servers, continue to the following steps.

5. Zip all `.bin` files created in steps 1 - 3 and place the zip archive on a web server.

    a. Insert the URL to your archive in the `garden.ini` configuration file as the `url` key in the `esp32` section.

6. Calculate the SHA256 hash of the archive and insert it into the `hash` key in the `esp32` section.

To test out your newly created archive:
  * Restart the garden backend server
  * Delete all files from the `data/firmware` directory
  * Click the Install Firmware button. The backend server should download and verify your archive successfully.

## Partition table
The included `partitions.bin` image contains the following partitions:
```
# ESP-IDF Partition Table
# Name, Type, SubType, Offset, Size, Flags
nvs,data,nvs,0x9000,20K,
otadata,data,ota,0xe000,8K,
app0,app,ota_0,0x10000,1280K,
app1,app,ota_1,0x150000,1280K,
spiffs,data,spiffs,0x290000,1472K,
```

The above CSV output can be recreated using [this](https://github.com/espressif/esp-idf/blob/4a011f3/components/partition_table/gen_esp32part.py) Python script from Espressif.