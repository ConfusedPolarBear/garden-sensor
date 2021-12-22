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
            system can connect to with access to the backend server.
          </p>

          <p>
            Systems that are updating will not communicate with the backend
            server or relay mesh messages for up to two minutes. In the event of
            an error, the system will restart itself without installing the
            update.
          </p>

          <v-form>
            <v-text-field v-model="update.ssid" label="SSID" />
            <v-text-field
              v-model="update.psk"
              label="Password"
              type="password"
            />
            <v-text-field v-model="update.host" label="Host" />
            <br />

            <v-btn
              @click="startUpdate"
              :disabled="this.update.loading"
              :loading="this.update.loading"
              color="primary darken-1"
            >
              Start update
            </v-btn>
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
    </v-expansion-panels>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import { MutationPayload } from "vuex";

import api from "@/plugins/api";
import SensorPreview from "@/components/SensorPreview.vue";
import { GardenSystem } from "@/store/types";

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
        loading: false
      }
    };
  },
  computed: {
    id() {
      return this.$route.params["id"];
    },
    chipset() {
      return this.$data.system?.Announcement?.Chipset;
    }
  },
  methods: {
    onMutation(mutation: MutationPayload) {
      if (
        mutation.type === "update" &&
        mutation.payload.Identifier === this.id
      ) {
        this.system = mutation.payload;
        /* TODO: fix infinite spinner by:
         *   creating an update log on this screen that defaults to "sent update command"
         *   the backend can use a new property called UpdateStatus that it updates when a system:
         *     publishes a message to tele/whatever/ota
         *     makes a request to /fwNN with a system ID header
         *     restarts with a new firmware version
         */
      }
    },
    startUpdate() {
      window.localStorage.setItem("updateSsid", this.update.ssid);
      window.localStorage.setItem("updatePass", this.update.psk);
      window.localStorage.setItem("updateHost", this.update.host);

      this.update.loading = true;

      api(`/system/update/${this.id}`, {
        method: "POST",
        body: new URLSearchParams({
          host: this.update.host,
          ssid: this.update.ssid,
          psk: this.update.psk
        })
      });
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
