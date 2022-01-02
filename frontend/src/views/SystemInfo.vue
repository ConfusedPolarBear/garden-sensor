<template>
  <v-container>
    <h2>{{ id }} ({{ chipset }})</h2>
    <v-container v-if="!initializing">
      <v-row justify="center" justify-sm="start">
        <v-col>
          <router-link :to="`/graph/${id}/Temperature`">
            <sensor-preview
              name="Temperature"
              icon="mdi-thermometer"
              :data="system.LastReading.Temperature"
            />
          </router-link>
        </v-col>
        <v-col>
          <router-link :to="`/graph/${id}/Humidity`">
            <sensor-preview
              name="Humidity"
              icon="mdi-water-percent"
              :data="system.LastReading.Humidity"
            />
          </router-link>
        </v-col>
      </v-row>
      <v-row>
        <v-col lg cols="6">
          <h2>Update firmware</h2>
          <p>
            To update this system, specify a Wi-Fi network and password the
            system can connect to with access to the backend server. In the
            event of an error, the system will attempt to restart itself without
            installing the update.
          </p>

          <span v-if="!update.hideWarning">
            <v-alert @input="hideWarning" dismissible outlined type="warning">
              <p class="mb-0">
                Systems that are updating will not communicate with the backend
                server or relay mesh messages for up to one minute.
              </p>

              <p class="mb-0">
                While the system will make a best effort attempt to
                automatically recover from any error that occurs during the
                update process, it is possible (but unlikely) that you may need
                to manually restart or reflash the system over a serial
                connection.
              </p>
            </v-alert>
            <br />
          </span>

          <v-form>
            <v-text-field v-model="update.ssid" label="Network name" />
            <v-text-field
              v-model="update.psk"
              label="Network password"
              type="password"
            />
            <v-text-field
              v-model="update.host"
              label="Backend server address"
            />

            <v-select
              v-if="update.showForceError"
              label="Inject error into"
              v-model="update.forceError"
              :items="update.forceErrors"
            />
            <br />

            <v-btn
              @click="startUpdate"
              :disabled="disallowUpdate"
              :loading="this.update.updating"
              :color="startButtonColor"
            >
              Start update
            </v-btn>

            <span>
              <br />
              <br />
              <span>Update event log:</span>
              <v-data-table
                :headers="this.update.eventHeaders"
                :items="this.update.events"
                :item-class="messageClass"
                no-data-text="No events yet"
              >
                <template v-slot:item.LoggedAt="{ item }">
                  {{ item.LoggedAt.toLocaleTimeString() }}
                  {{ item.Delta }}
                </template>
              </v-data-table>
            </span>
          </v-form>
        </v-col>
      </v-row>
    </v-container>
    <br />

    <v-expansion-panels>
      <v-expansion-panel>
        <v-expansion-panel-header>
          I want to update a system from a network without access to my backend
          server
        </v-expansion-panel-header>
        <v-expansion-panel-content>
          <span>
            It is possible to update systems even if they are on a separate
            network from the backend server. To download a firmware update,
            systems make an HTTP GET request to either:
          </span>

          <ul>
            <li>
              <code>http://UPDATE_HOST/fw82</code> is the system is running on
              an ESP8266, or
            </li>
            <li>
              <code>http://UPDATE_HOST/fw32</code> is the system is running on
              an ESP32
            </li>
          </ul>
          <br />

          <span>
            Your HTTP server must:
            <ol>
              <li>
                Allow unauthenticated HTTP access to <code>/fw82</code> and
                <code>/fw32</code>
              </li>
              <li>Send a correct Content-Length header</li>
            </ol>
            <br />

            If you have <code>python3</code> installed on your system, the
            command <code>python3 -m http.server</code> works well.
          </span>
        </v-expansion-panel-content>
      </v-expansion-panel>

      <v-expansion-panel>
        <v-expansion-panel-header>Debug updater</v-expansion-panel-header>
        <v-expansion-panel-content>
          <p>
            Various errors can be artifically injected into the update process
            for debugging purposes.
          </p>

          <v-btn @click="toggleErrorInjection" color="error darken-1">
            Toggle error injection
          </v-btn>
        </v-expansion-panel-content>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import { MutationPayload } from "vuex";

import api from "@/plugins/api";
import SensorPreview from "@/components/SensorPreview.vue";
import { GardenSystem, OTAStatus } from "@/store/types";

export default Vue.extend({
  name: "SystemInfo",
  components: { SensorPreview },
  data() {
    return {
      initializing: true,
      system: {} as GardenSystem,
      update: {
        ssid: window.localStorage.getItem("updateSsid") || "",
        psk: window.localStorage.getItem("updatePass") || "",
        host: window.localStorage.getItem("updateHost") || window.location.host,
        updating: false,

        hideWarning:
          window.localStorage.getItem("updateShowWarning") === "false",

        showForceError: window.localStorage.getItem("updateError") === "true",
        forceError: "",
        forceErrors: [
          { text: "Nothing", value: "" },
          { text: "SSID (randomize)", value: "ssid" },
          { text: "PSK (randomize)", value: "psk" },
          { text: "Host (set to localhost)", value: "host" },
          { text: "URL (set to /dead)", value: "url" },
          { text: "Firmware size (half of original value)", value: "size" },
          { text: "Checksum (randomize)", value: "checksum" }
        ],

        lastEvent: {} as OTAStatus,
        events: Array<OTAStatus>(),
        eventHeaders: [
          {
            text: "Time",
            value: "LoggedAt",
            width: "15%"
          },
          {
            text: "Message",
            value: "Message"
          }
        ]
      }
    };
  },
  computed: {
    id(): string {
      return this.$route.params["id"];
    },
    chipset(): string {
      return this.$data.system?.Announcement?.Chipset;
    },
    disallowUpdate(): boolean {
      if (this.update.updating) {
        return true;
      }

      const u = this.$data.update;
      return !u.ssid || !u.psk || !u.host;
    },
    startButtonColor(): string {
      return this.update.forceError ? "red darken-1" : "primary darken-1";
    }
  },
  methods: {
    onMutation(mutation: MutationPayload) {
      if (
        mutation.type === "update" &&
        mutation.payload.Identifier === this.id
      ) {
        this.system = mutation.payload;

        // See if the system sent a new update status message
        const update = this.system.UpdateStatus;
        if (!update || !update.Message) {
          return;
        }

        let last = this.$data.update.lastEvent;
        if (last.Message === update.Message) {
          return;
        }

        if (!update.Message.startsWith("Backend")) {
          update.Message = `System: ${update.Message}`;
        }

        this.pushEvent(update);

        if (!update.Success || update.Message.indexOf("restart") !== -1) {
          this.$data.update.updating = false;
        }
      }
    },
    startUpdate() {
      this.$data.update.events.splice(0);

      window.localStorage.setItem("updateSsid", this.update.ssid);
      window.localStorage.setItem("updatePass", this.update.psk);
      window.localStorage.setItem("updateHost", this.update.host);

      this.update.updating = true;

      api(`/system/update/${this.id}`, {
        method: "POST",
        body: new URLSearchParams({
          host: this.update.host,
          ssid: this.update.ssid,
          psk: this.update.psk,
          error: this.update.forceError
        })
      });

      this.pushEvent({
        Success: true,
        Message: `Backend: Sent update command to ${this.id}`
      });
    },
    pushEvent(e: OTAStatus): void {
      e.LoggedAt = new Date();
      e.Delta = this.getElapsed();
      this.$data.update.events.push(e);
      this.$data.update.lastEvent = e;
    },
    messageClass(row: OTAStatus): string {
      return row.Success ? "" : "red";
    },
    getElapsed(): string {
      const rhs = this.$data.update.lastEvent.LoggedAt;
      const d = (Number(new Date()) - rhs) / 1000;

      if (isNaN(d) || d <= 0.05) {
        return "";
      }

      return `(+${d.toFixed(2)} sec)`;
    },
    hideWarning(): void {
      console.debug("hi");
      window.localStorage.setItem("updateShowWarning", "false");
      this.$data.update.hideWarning = true;
    },
    toggleErrorInjection() {
      const key = "updateError";
      if (window.localStorage.getItem(key)) {
        window.localStorage.removeItem(key);
      } else {
        window.localStorage.setItem(key, "true");
      }

      let err = this.update.showForceError;
      if (err) {
        this.update.forceError = "";
      }
      this.update.showForceError = !err;
    }
  },
  async created() {
    // TODO: retrieve these from vuex instead of this separate API call
    const res = await api(`/system/${this.$route.params.id}`);
    this.system = await res.json();
    this.initializing = false;
  },
  mounted() {
    this.$store.subscribe(this.onMutation);
  }
});
</script>

<style scoped>
a {
  text-decoration: none;
}
.col {
  flex-grow: 0;
}
</style>
