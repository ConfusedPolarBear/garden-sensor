<template>
  <v-container>
    <p>Systems:</p>
    <v-data-table :items="$store.state.systems" :headers="headers">
      <template v-slot:[`item.Identifier`]="{ item }">
        <router-link :to="`/system/${item.Identifier}`">
          <code>{{ item.Identifier }}</code>
        </router-link>
      </template>

      <template v-slot:[`item.Connection`]="{ item }">
        <div class="readingData">
          <tooltip :text="meshInfo(item, 'tooltip')">
            <v-icon>
              {{ meshInfo(item, "icon") }}
            </v-icon>
          </tooltip>

          <span v-if="showSystemTypes" style="margin-left: 1rem">
            <tooltip text="Virtual">
              <v-icon v-if="isEmulator(item)">mdi-progress-wrench</v-icon>
            </tooltip>

            <tooltip text="Physical">
              <v-icon v-if="!isEmulator(item)">mdi-memory</v-icon>
            </tooltip>
          </span>
        </div>
      </template>

      <template v-slot:[`item.LastReading`]="{ item }">
        <div id="reading" v-if="dataValid(item.UpdatedAt)">
          <tooltip
            v-if="item.LastReading.Error"
            text="Error retrieving sensor data"
          >
            <v-icon color="red">mdi-alert</v-icon>
          </tooltip>

          <div v-if="!item.LastReading.Error">
            <div
              class="readingData"
              v-if="isReadingValid(item.LastReading.Temperature)"
            >
              <tooltip text="Temperature">
                <v-icon>mdi-thermometer</v-icon>
                <span>{{ temp(item.LastReading.Temperature) }}</span>
              </tooltip>
            </div>

            <div
              class="readingData"
              v-if="isReadingValid(item.LastReading.Humidity)"
            >
              <tooltip text="Humidity">
                <v-icon>mdi-water-percent</v-icon>
                <span>{{ item.LastReading.Humidity.toFixed(2) }} %</span>
              </tooltip>
            </div>
          </div>
          <div class="readingData" v-else>
            <span class="red--text">Error retrieving sensor data</span>
          </div>
        </div>
        <div id="reading" v-else>
          <v-icon>mdi-clock</v-icon>
          <span>has not reported data yet</span>
        </div>
      </template>

      <template v-slot:[`item.UpdatedAt`]="{ item }">
        <div class="readingData">
          <v-icon style="margin-right: 0.5rem">mdi-clock</v-icon>
          <span>{{ age(item.UpdatedAt) }}</span>
        </div>
      </template>

      <template v-slot:[`item.Filesystem`]="{ item }">
        <span v-if="!isEmulator(item)">
          <div class="readingData">
            <v-icon>mdi-file-multiple</v-icon>
            {{ fsInfo(item.Announcement) }}
          </div>
        </span>
      </template>
    </v-data-table>
    <br />

    <span>Send command to:</span>
    <br />
    <span v-for="sys in $store.state.systems" :key="sys.Identifier">
      <a href="#" @click="sendCommand(sys.Identifier)">
        <code>{{ sys.Identifier }}</code>
        <br />
      </a>
    </span>
    <a href="#" @click="sendCommand('FFFFFFFFFFFF')">
      <code>All systems</code>
      <br />
    </a>

    <v-btn
      @click="$router.push('/configure')"
      elevation="2"
      fab
      color="green"
      style="position: relative; top: 5em; float: right"
    >
      <v-icon>mdi-plus</v-icon>
    </v-btn>

    <br />
    <div style="max-width: 300px">
      <h2>Temporary settings menu</h2>
      <v-switch v-model="fahrenheit" label="Use Farenheit" />
    </div>

    <command-dialog
      @command="sendCommandHandler"
      @close="command.show = false"
      :id="command.id"
      :show="command.show"
    />
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import api from "@/plugins/api";
import { GardenSystem, GardenSystemInfo } from "@/store/types";
import { MutationPayload } from "vuex";

import CommandDialog from "@/components/CommandDialog.vue";
import Tooltip from "@/components/Tooltip.vue";

export default Vue.extend({
  name: "Systems",
  components: { CommandDialog, Tooltip },
  data() {
    return {
      headers: [
        {
          text: "Identifier",
          value: "Identifier",
          width: "10%"
        },
        {
          text: "Connection",
          value: "Connection",
          width: "10%"
        },
        {
          text: "Last Reading",
          value: "LastReading",
          width: "25%"
        },
        {
          text: "Last Seen",
          value: "UpdatedAt"
        },
        {
          text: "Filesystem",
          value: "Filesystem"
        }
      ],
      systems: Array<GardenSystem>(),
      showSystemTypes: false,
      fahrenheit: (window.localStorage.getItem("units") ?? "C") == "F",

      command: {
        id: "",
        show: false
      }
    };
  },
  methods: {
    load(): void {
      // Load all initial systems.
      api("/systems");

      // Subscribe to the Vuex store for future updates.
      this.$store.subscribe(this.onMutation);
    },
    onMutation(mutation: MutationPayload, state: any) {
      if (mutation.type !== "register" && mutation.type !== "update") {
        return;
      }

      this.systems = state.systems;

      // When a new system is registered, check if any emulators are present. If there are any, display system types.
      // Don't bother to check if this isn't a new system registering itself or if we are already showing type info.
      if (mutation.type !== "register" || this.showSystemTypes) {
        return;
      }

      this.showSystemTypes = false;
      for (const sys of this.systems) {
        if (this.isEmulator(sys)) {
          this.showSystemTypes = true;
          break;
        }
      }
    },
    dataValid(lastSeen: string): boolean {
      return !lastSeen.startsWith("0001");
    },
    age(lastSeen: string): string {
      let diff = Number(new Date()) - Number(new Date(lastSeen));
      diff = Number(diff) / 1000;
      return `last seen ${diff.toFixed(0)} seconds ago`;
    },
    isEmulator(system: GardenSystem): boolean {
      return system.Announcement.IsEmulator;
    },
    fsInfo(info: GardenSystemInfo): string {
      const used = info.FilesystemUsedSize / 1024;
      const total = info.FilesystemTotalSize / 1024;

      if (total === 0) {
        return "No filesystem present";
      }

      const percent = ((used * 100) / total).toFixed(2);

      return `${used}K (${percent}%) used out of ${total}K total`;
    },
    meshInfo(system: GardenSystem, item: string): any {
      const mesh = system.Announcement.IsMesh;

      switch (item) {
        case "tooltip":
          return mesh ? "Mesh" : `MQTT (CH ${system.Announcement.Channel})`;

        case "icon":
          return mesh ? "mdi-access-point" : "mdi-wifi";

        default:
          throw new Error(`unknown meshInfo item ${item}`);
      }
    },
    isReadingValid(reading: number): boolean {
      return reading != 32768;
    },
    temp(reading: number): string {
      const units = this.fahrenheit ? "F" : "C";

      if (this.fahrenheit) {
        reading = (9 / 5) * reading + 32;
      }

      return `${reading.toFixed(2)} °${units}`;
    },

    sendCommand(id: string) {
      this.command.id = id;
      this.command.show = true;
    },
    sendCommandHandler(data: any) {
      const id = data.id;
      const command = data.command;

      api(`/system/command/${id}`, {
        method: "POST",
        body: new URLSearchParams({
          command: command
        })
      });
    }
  },
  created() {
    this.load();

    // Periodically force the page to update in order to keep the last seen timestamps fresh for all systems.
    setInterval(() => {
      this.$forceUpdate();
    }, 5 * 1000);
  },
  watch: {
    fahrenheit() {
      window.localStorage.setItem("units", this.fahrenheit ? "F" : "C");
    }
  }
});
</script>

<style scoped>
div#reading span {
  margin-left: 0.1rem;
  margin-right: 0.1rem;
}

/* This ensures the reading icons aren't separated from the reading data point when wrapped on mobile */
@media screen and (min-width: 1000px) {
  div.readingData {
    display: inline;
  }
}
</style>
