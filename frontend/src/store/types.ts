export interface StoreState {
  systems: Array<GardenSystem>;
}

// Copied from /backend/internal/util/

export type Sensor = {
  GardenSystemInfoID: string;
  Name: string;
};

export type GardenSystem = {
  Identifier: string;
  CreatedAt: Date;
  UpdatedAt: Date;
  DeletedAt: Date;
  Announcement: GardenSystemInfo;
  LastReading: Reading;
  Readings: Array<Reading>;
};

export type GardenSystemInfo = {
  // Parent garden system that generated this announcement.
  GardenSystemID: string;

  // If this garden system is an actually an emulator. This field should not be sent by non-virtual systems.
  IsEmulator: boolean;

  // If this system is connected through the mesh or MQTT.
  IsMesh: boolean;

  // Channel that the Wi-Fi station uses. Only valid if this is a mesh controller.
  Channel: number;

  RestartReason: string;

  // Chipset on this system. Either "ESP8266" or "ESP32".
  Chipset: string;
  CoreVersion: string;
  SdkVersion: string;

  FilesystemUsedSize: number;
  FilesystemTotalSize: number;

  Sensors: Array<Sensor>;
};

export type Reading = {
  CreatedAt: Date;
  // Parent garden system that generated this reading.
  GardenSystemID: string;
  Error: boolean;
  Temperature?: number;
  Humidity?: number;
};
