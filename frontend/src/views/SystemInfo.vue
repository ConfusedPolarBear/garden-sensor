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
            Systems can only be updated over Wi-Fi. Specify the network name and
            password of a network that this system should connect to in order to
            download the new firmware. You must also specify the IP address and
            port of the backend server on that network.
          </p>

          <p>
            Systems that are downloading new firmware will disconnect from MQTT
            and the mesh for up to two minutes. In the event of an error, the
            system will restart itself without installing the new firmware.
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
    <v-expansion-panels>
      <v-expansion-panel>
        <v-expansion-panel-header>Debug Info</v-expansion-panel-header>
        <v-expansion-panel-content>
          <pre> {{ JSON.stringify(system, undefined, 2) }} </pre>
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
        this.update.loading = false;
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
