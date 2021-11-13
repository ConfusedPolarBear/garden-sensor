<template>
  <v-container>
    <v-stepper v-model="step">
      <v-stepper-header>
        <v-stepper-step :complete="step > 1" step="1">
          Install firmware
        </v-stepper-step>

        <v-stepper-step :complete="step > 2" step="2">
          Select networking mode
        </v-stepper-step>

        <v-stepper-step :complete="step > 3" step="3">
          Upload configuration
        </v-stepper-step>

        <v-stepper-step :complete="step > 4" step="4">
          Complete
        </v-stepper-step>
      </v-stepper-header>

      <v-stepper-items>
        <v-stepper-content step="1">
          <p>
            Install (or "flash") the latest stable garden firmware to your
            ESP8266 or ESP32.
          </p>

          <p>
            Note: if you are flashing a ESP32 chip and you have not flashed an
            ESP32 chip before, you will need to download
            <a href="https://example.com" target="_blank">this ZIP file</a>
            and place it into your
            <code>data</code>
            directory.
          </p>

          <v-alert color="orange darken-2">TODO: fix ESP32 blob link</v-alert>

          <p>Once you have successfully installed the firmware, click Next.</p>

          <esp-web-install-button :manifest="manifestUrl">
            <v-btn slot="activate" color="success darken-1">
              Install firmware
            </v-btn>
            <span slot="unsupported">
              Your browser does not support the required APIs for in-browser
              flashing. You can still flash manually with esptool.
            </span>
            <span slot="not-allowed">
              This page must be loaded over HTTPS in order to use in-browser
              flashing. You can still flash manually with esptool.
            </span>
          </esp-web-install-button>
          <br />
          <br />

          <v-expansion-panels outlined>
            <v-expansion-panel>
              <v-expansion-panel-header>
                Manual installation instructions
              </v-expansion-panel-header>

              <v-expansion-panel-content>
                <p>
                  If you are unable to use the in-browser flashing tool, follow
                  the below instructions to successfully flash the firmware onto
                  your garden system.
                </p>

                <ol>
                  <li>Install <code>esptool</code>.</li>
                  <li>Download the latest stable firmware binary.</li>
                  <li>Open a terminal window (or command prompt).</li>
                  <li>
                    If this chip is an ESP8266, run
                    <code>esptool write_flash 0x0 firmware.bin</code>
                  </li>
                  <li>
                    If this chip is an ESP32, run
                    <code>
                      esptool write_flash 0x1000 bootloader.bin 0x8000
                      partitions.bin 0xe000 boot_app0.bin 0x10000 firmware.bin
                    </code>
                  </li>
                </ol>
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
          <br />

          <v-btn color="primary" @click="step++"> Next &gt; </v-btn>
        </v-stepper-content>

        <v-stepper-content step="2">
          <p>
            Garden systems can either be connected through Wi-Fi or a mesh
            network. There must be at least one system connected through Wi-Fi
            in each mesh network.
            <br />
            Choose a networking mode to continue.
          </p>

          <div class="center">
            <wizard-card
              type="wifi"
              @click="typeSelected"
              :selected="systemType"
            />

            <wizard-card
              type="mesh"
              @click="typeSelected"
              :selected="systemType"
            />
          </div>
          <br />

          <div v-if="systemType === 'wifi'">
            <p>
              This system will be connected to a Wi-Fi network and use MQTT to
              communicate with the server. It will also act as a mesh controller
              and permit future systems to communicate directly with it.
            </p>

            <tooltip text="Recommended for grid powered systems">
              <v-icon large>mdi-power-plug</v-icon>
            </tooltip>
          </div>
          <div v-else-if="systemType === 'mesh'">
            <p>
              This system will be connected to a mesh network created by your
              current garden systems.
            </p>

            <tooltip text="Recommended for battery powered systems">
              <v-icon large>mdi-battery</v-icon>
            </tooltip>
          </div>
          <br />

          <v-btn color="primary" @click="step++" :disabled="systemType === ''">
            Next &gt;
          </v-btn>
        </v-stepper-content>

        <v-stepper-content step="3">
          <v-form v-if="systemType == 'wifi'">
            <v-text-field
              v-model="config.wifiSsid"
              label="Wi-Fi SSID"
              hint="Name of your Wi-Fi network."
            />

            <v-text-field
              v-model="config.wifiPass"
              type="password"
              max-length="64"
              label="Wi-Fi password"
              hint="Password for your Wi-Fi network."
            />

            <v-text-field
              v-model="config.mqttHost"
              label="MQTT Address"
              hint="Address of your MQTT broker."
            />

            <v-text-field
              v-model="config.mqttUser"
              label="MQTT username (optional)"
              hint="If your broker requires authentication, username to authenticate with."
            />

            <v-text-field
              v-model="config.mqttPass"
              type="password"
              label="MQTT password (optional)"
              hint="If your broker requires authentication, password to authenticate with."
            />
          </v-form>

          <v-form v-else-if="systemType === 'mesh'">
            <v-text-field
              v-model="config.meshController"
              label="Controller address"
              hint="The MAC address of the mesh controller."
            />

            <v-text-field
              v-model="config.meshChannel"
              type="number"
              min="1"
              max="14"
              label="Channel"
              hint="The channel that the mesh controller is listening on."
            />

            <v-alert color="warning darken-1">
              TODO: Both of these fields need to be autofilled by the server and
              displayed as dropdowns.
            </v-alert>
          </v-form>

          <v-btn
            @click="saveConfig"
            :disabled="!configValid"
            color="blue darken-1"
          >
            Save and upload
          </v-btn>

          <p style="margin-top: 10px">
            Configuration details are saved to this computer and never
            transmitted anywhere except your garden systems.
          </p>

          <v-expansion-panels v-model="configPanel">
            <v-expansion-panel v-model="configPanel">
              <v-expansion-panel-header>
                Advanced configuration upload
              </v-expansion-panel-header>

              <v-expansion-panel-content>
                <p>
                  If the WebSerial API is not supported in your browser, you
                  will need to upload your configuration manually. To do so:
                </p>
                <ul>
                  <li>
                    Click the clipboard icon below to copy the JSON
                    configuration to your clipboard.
                  </li>
                  <li>
                    Open a serial monitor on your computer and paste the
                    generated JSON configuration.
                  </li>
                  <li>Restart the ESP.</li>
                </ul>
                <br />

                <v-icon @click="copy()" v-show="secure" id="copy">
                  mdi-clipboard-multiple-outline
                </v-icon>

                <code id="jsonConfig">{{ this.serializeConfig() }}</code>
                <br />
                <br />

                <v-btn color="primary" @click="step++">Next &gt;</v-btn>
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-stepper-content>

        <v-stepper-content step="4">
          <p>
            Your new garden system has been successfully flashed and configured.
            You can either go back to
            <router-link to="/">the dashboard</router-link>
            to see your new system in action or you can
            <a href="#" @click="step = 1">return to step 1</a>
            to configure another system.
          </p>
        </v-stepper-content>
      </v-stepper-items>
    </v-stepper>

    <br />
    <v-btn color="primary" @click="goBack" v-show="step > 1">&lt; Back</v-btn>

    <v-snackbar v-model="snackbar.show" :color="snackbar.color" timeout="3000">
      <strong>{{ snackbar.text }}</strong>
    </v-snackbar>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import WizardCard from "@/components/WizardCard.vue";
import Tooltip from "@/components/Tooltip.vue";

import "esp-web-tools";

import { Dictionary } from "vue-router/types/router";

export default Vue.extend({
  name: "SystemWizard",
  components: { WizardCard, Tooltip },
  data() {
    return {
      step: 1,
      systemType: "",
      configPanel: -1,

      config: {
        wifiSsid: window.localStorage.getItem("wifiSsid") || "",
        wifiPass: window.localStorage.getItem("wifiPass") || "",
        mqttHost: window.localStorage.getItem("mqttHost") || "",
        mqttUser: window.localStorage.getItem("mqttUser") || "",
        mqttPass: window.localStorage.getItem("mqttPass") || "",
        meshController: window.localStorage.getItem("meshController") || "",
        meshChannel: window.localStorage.getItem("meshChannel") || ""
      } as Dictionary<string>,

      snackbar: {
        color: "green",
        show: false,
        text: ""
      }
    };
  },
  computed: {
    secure() {
      return window.isSecureContext;
    },
    webSerialSupported() {
      return "serial" in navigator;
    },
    manifestUrl(): string {
      return window.localStorage.getItem("server") + "/firmware/manifest.json";
    },
    configValid(): boolean {
      const s = (d: string): boolean => {
        return d.length > 0;
      };
      const c = this.config;

      switch (this.$data.systemType) {
        case "wifi":
          return s(c.wifiSsid) && s(c.wifiPass) && s(c.mqttHost);

        case "mesh":
          return s(c.meshController) && s(c.meshChannel);

        default:
          return false;
      }
    }
  },
  methods: {
    typeSelected(type: string): void {
      this.systemType = type;
    },

    // Configuration (de)serialization
    saveConfig() {
      for (const key in this.config) {
        window.localStorage.setItem(key, this.config[key]);
      }

      this.showSnackbar("Configuration saved locally");

      if (this.webSerialSupported) {
        this.uploadConfiguration();
      } else {
        this.showSnackbar(
          "WebSerial API is not supported, view advanced configuration for details.",
          "red"
        );

        this.configPanel = 0;
      }
    },
    serializeConfig(): string {
      let serialized = {} as any;

      switch (this.$data.systemType) {
        case "wifi":
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

          break;

        case "mesh":
          serialized.MeshController = this.$data.config.meshController;
          serialized.MeshChannel = this.$data.config.meshChannel;

          break;
      }

      return JSON.stringify(serialized);
    },

    async uploadConfiguration() {
      try {
        console.debug("requesting port");
        const port = await navigator.serial.requestPort();

        console.debug("opening port");
        await port.open({ baudRate: 115200 });

        const config = this.serializeConfig();
        console.debug("opened port, writing configuration", config);

        const writer = port.writable.getWriter();
        await writer.write(this.stringToBytes(config + "\n"));
        await writer.write(this.stringToBytes(`{"Command":"Restart"}\n`));

        console.debug("write done, closing port");
        await writer.releaseLock();
        await port.close();

        this.showSnackbar("Configuration successfully uploaded");
        this.step++;
      } catch (e) {
        this.showSnackbar("Unable to open serial port.", "red");
        console.error(e);
      }
    },

    // Encodes the provided string into a UTF-8 byte array.
    stringToBytes(raw: string): Uint8Array {
      return new TextEncoder().encode(raw);
    },

    // Utility functions
    goBack() {
      this.step = Math.max(1, this.step - 1);
    },
    showSnackbar(text: string, color = "green") {
      this.snackbar = {
        color: color,
        show: true,
        text: text
      };
    },
    copy(): void {
      const text = this.serializeConfig();
      navigator.clipboard.writeText(text).then(
        () => {
          this.showSnackbar("Copied!");
        },
        () => {
          this.showSnackbar("Failed to copy configuration", "red");
        }
      );
    }
  }
});
</script>

<style scoped>
.center {
  display: flex;
  justify-content: center;
}

#copy {
  margin-right: 10px;
}

#jsonConfig {
  background-color: transparent;
  border: 1px dashed gray;
  word-wrap: break-word;
}

/* Add a small margin between form inputs */
div[class*="v-input"] {
  margin-bottom: 10px;
}
</style>
