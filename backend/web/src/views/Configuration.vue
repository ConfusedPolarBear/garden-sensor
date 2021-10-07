<template>
  <v-container>
    <div v-if="secure">
      <!-- Port list -->
      <p>Available ports are listed below, click on one to connect to it.</p>
      <ul v-for="port in ports" :key="getVidPid(port)">
        <li>
          <a href="#" class="serialPort" @click.prevent="connect(port)">
            {{ getVidPid(port) }}
          </a>
        </li>
      </ul>
      <br />

      <div v-if="connected">
        <p>Connected to device</p>

        <!-- Serial monitor -->
        <v-textarea v-model="text" rows="10" readonly id="output"></v-textarea>
        <v-text-field
          @click:append="sendCommand"
          @keydown="commandKeyDown"
          v-model="command"
          label="Send text"
          append-icon="mdi-chevron-right"
        />

        <v-checkbox
          v-model="autoscroll"
          label="Autoscroll output to bottom"
        ></v-checkbox>

        <!-- Toolbar -->
        <span v-for="button in buttons" :key="button.text">
          <v-btn
            @click="button.action"
            :color="button.color"
            style="margin-right: 5px"
          >
            {{ button.text }}
          </v-btn>
        </span>
      </div>
      <div v-else>
        <!-- Serial port connection -->
        <v-btn @click="requestPort" color="blue">Connect to Device</v-btn>

        <v-checkbox
          v-model="useVendorList"
          label="Limit connections to known serial adapters"
        >
        </v-checkbox>
        <p v-if="!useVendorList">
          If you have a device which does not appear with the above checkbox
          checked, please open an issue on our repository with the vendor ID.
        </p>
      </div>
    </div>
    <v-banner v-else color="red">
      <span>
        The browser serial monitor only works in browsers which support
        WebSerial and only when accessed from localhost or over HTTPS. <br />
        Both of these limitations are imposed by your browser - not this
        software.
      </span>
    </v-banner>

    <!-- Configuration editor -->
    <v-form style="margin-top: 2em; max-width: 50%">
      <h2>Configuration editor</h2>

      <v-text-field v-model="config.wifiSsid" label="Wi-Fi SSID" />
      <v-text-field
        v-model="config.wifiPass"
        type="password"
        max-length="64"
        label="Wi-Fi password"
      />

      <v-text-field v-model="config.mqttHost" label="MQTT Address" />
      <v-text-field
        v-model="config.mqttUser"
        label="MQTT username (optional)"
      />
      <v-text-field
        v-model="config.mqttPass"
        type="password"
        label="MQTT password (optional)"
      />

      <v-btn @click="saveConfig" color="blue darken-1">Save changes</v-btn>
      <p style="margin-top: 10px">
        Configuration details are saved to this computer and never transmitted
        anywhere except your garden systems.
      </p>

      <p>
        Configuration details have been serialized to JSON below. The text has
        been blurred to protect your credentials. Hovering with the mouse will
        remove the blur.
      </p>
      <pre id="configJson">{{ this.serializeConfig(true) }}</pre>
    </v-form>

    <v-snackbar v-model="snackbar.show" :color="snackbar.color" timeout="3000">
      <strong>{{ snackbar.text }}</strong>
    </v-snackbar>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import { Dictionary } from "vue-router/types/router";

export default Vue.extend({
  name: "Configuration",
  data() {
    return {
      secure: self.isSecureContext && navigator.serial,
      ports: Array<unknown>(),
      text: "",
      command: "",

      // Default allowlist of vendor IDs to filter by.
      knownVendors: [
        // Appears as "USB2.0 Ser!". Chipset: CH341. Found on WeMos D1 Mini clones.
        0x1a86
      ],
      useVendorList: true,

      interval: 0,
      connected: false,
      autoscroll: true,
      writer: {} as WritableStream,
      config: {
        wifiSsid: window.localStorage.getItem("wifiSsid") || "",
        wifiPass: window.localStorage.getItem("wifiPass") || "",
        mqttHost: window.localStorage.getItem("mqttHost") || "",
        mqttUser: window.localStorage.getItem("mqttUser") || "",
        mqttPass: window.localStorage.getItem("mqttPass") || ""
      } as Dictionary<string>,

      // TODO: fix the "errors" on the button action
      buttons: [
        {
          text: "Get Info",
          color: "green darken-2",
          action: this.getInfo
        },
        {
          text: "Update Configuration",
          color: "blue darken-2",
          action: this.updateConfig
        },
        {
          text: "Reboot",
          color: "orange darken-1",
          action: this.restart
        },
        {
          text: "Factory Reset",
          color: "red darken-2",
          action: this.factoryReset
        }
      ],

      snackbar: {
        color: "green",
        show: false,
        text: ""
      }
    };
  },
  methods: {
    // High level serial stuff
    async listPorts() {
      try {
        // eslint-ignore-next-line
        this.ports = await navigator.serial.getPorts();
      } catch (e) {
        console.error("unable to list ports", e);
      }
    },
    async requestPort() {
      const filters = Array<unknown>();
      if (this.useVendorList) {
        console.debug("applying VID filters");
        for (const vid of this.knownVendors) {
          filters.push({ usbVendorId: vid });
        }
      } else {
        console.debug("not applying VID filters");
      }

      const port = await navigator.serial.requestPort({ filters: filters });
      this.listPorts();

      this.connect(port);
    },
    async connect(port: any) {
      await port.open({ baudRate: 115200, bufferSize: 10 * 1024 });

      this.connected = true;

      const read: ReadableStream = port.readable;

      // TODO: look into improvements here
      const reader = read.getReader();
      this.interval = setInterval(async () => {
        try {
          const resp = await reader.read();
          this.text += this.decode(resp.value);
        } catch (e) {
          console.error(e);
          clearInterval(this.interval);
        }
      }, 100);

      this.writer = port.writable;

      setTimeout(this.getInfo, 1000);
    },

    // Serial monitor
    sendCommand() {
      this.writeToPort(this.command);
      this.command = "";
    },
    commandKeyDown(e: KeyboardEvent) {
      if (e.code == "Enter") {
        this.sendCommand();
      }
    },

    // Configuration (de)serialization
    saveConfig() {
      for (const key in this.config) {
        window.localStorage.setItem(key, this.config[key]);
      }

      this.showSnackbar("Configuration saved locally");
    },
    serializeConfig(prettify = false): string {
      let serialized = {} as any;

      // Only update the Wi-Fi data if both the SSID and PSK are present.
      if (this.$data.config.wifiSsid && this.$data.config.wifiPass) {
        serialized.WifiSSID = this.$data.config.wifiSsid;
        serialized.WifiPassword = this.$data.config.wifiPass;
      }

      if (this.$data.config.mqttHost) {
        serialized.MQTTHost = this.$data.config.mqttHost;

        if (this.$data.config.mqttUser) {
          serialized.MQTTUsername = this.$data.config.mqttUser;
          serialized.MQTTPassword = this.$data.config.mqttPass;
        }
      }

      return JSON.stringify(serialized, null, prettify ? 4 : 0);
    },

    // Toolbar buttons
    getInfo() {
      this.writeToPort(`{"Command":"info"}`);
    },
    updateConfig() {
      this.writeToPort(this.serializeConfig());
      setTimeout(this.restart, 1000);
    },
    restart() {
      this.writeToPort(`{"Command":"restart"}`);
    },
    factoryReset() {
      // TODO: this needs a confirmation dialog
      this.writeToPort(`{"Command":"reset"}`);
    },

    // Serial port management
    async writeToPort(msg: string) {
      const w = this.writer.getWriter();
      await w.write(this.encode(msg));
      w.releaseLock();
    },
    encode(raw: string): Uint8Array {
      return new TextEncoder().encode(raw);
    },
    decode(raw: Uint8Array): string {
      return new TextDecoder().decode(raw);
    },
    getVidPid(port: any) {
      const i = port.getInfo();
      return i.usbVendorId.toString(16) + ":" + i.usbProductId.toString(16);
    },

    // Snackbar handling
    showSnackbar(text: string, color = "green") {
      this.snackbar = {
        color: color,
        show: true,
        text: text
      };
    }
  },
  mounted() {
    if (!this.secure) {
      return;
    }

    this.listPorts();
  },
  watch: {
    text() {
      const textarea = document.getElementById("output");
      if (!textarea || !this.autoscroll) {
        return;
      }

      textarea.scrollTop = 2 * textarea.scrollHeight;
    }
  }
});
</script>

<style>
#output {
  font-family: "monospace";
}

#configJson {
  color: transparent;
  text-shadow: 0 0 10px white;
}

#configJson:hover {
  color: white;
  text-shadow: unset;
}
</style>
