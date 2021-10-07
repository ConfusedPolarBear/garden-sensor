<template>
  <v-container>
    <p>Systems:</p>
    <v-data-table :items="$store.state.systems" :headers="headers">
      <template v-slot:[`item.Identifier`]="{ item }">
        <code>{{ item.Identifier }}</code>
      </template>

      <template v-slot:[`item.LastReading`]="{ item }">
        <div id="reading" v-if="dataValid(item.LastSeen)">
          <span v-if="showSystemTypes">
            <v-icon v-if="isEmulator(item)">mdi-progress-wrench</v-icon>
            <v-icon v-else>mdi-memory</v-icon>
          </span>

          <div class="readingData">
            <v-icon>mdi-thermometer</v-icon>
            <span>{{ item.LastReading.Temperature }} °C</span>
          </div>

          <div class="readingData">
            <v-icon>mdi-water-percent</v-icon>
            <span>{{ item.LastReading.Humidity }} %</span>
          </div>

          <div class="readingData">
            <v-icon>mdi-clock</v-icon>
            <span>{{ age(item.LastSeen) }}</span>
          </div>

          <span v-if="!isEmulator(item)">
            <div class="readingData">
              <v-icon>mdi-file-multiple</v-icon>
              {{ fsInfo(item.Announcement.System) }}
            </div>
          </span>
        </div>
        <div id="reading" v-else>
          <v-icon>mdi-clock</v-icon>
          <span>has not reported data yet</span>
        </div>
      </template>
    </v-data-table>

    <v-btn
      @click="$router.push('/configure')"
      elevation="2"
      fab
      color="green"
      style="position: relative; top: 5em; float: right"
    >
      <v-icon>mdi-plus</v-icon>
    </v-btn>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import api from "@/plugins/api";
import { MutationPayload } from "vuex";

export default Vue.extend({
  name: "Systems",
  data() {
    return {
      headers: [
        {
          text: "Identifier",
          value: "Identifier",
          width: "10%"
        },
        {
          text: "Last Reading",
          value: "LastReading"
        }
      ],
      systems: Array<unknown>(),
      showSystemTypes: false
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
    isEmulator(system: any): boolean {
      return system.Announcement.System.IsEmulator;
    },
    fsInfo(info: any): string {
      const used = info.FilesystemUsedSize / 1024;
      const total = info.FilesystemTotalSize / 1024;

      if (total === 0) {
        return "No filesystem present";
      }

      const percent = ((used * 100) / total).toFixed(2);

      return `${used}K (${percent}%) used out of ${total}K total`;
    }
  },
  created() {
    this.load();

    // Periodically force the page to update in order to keep the last seen timestamps fresh for all systems.
    setInterval(() => {
      this.$forceUpdate();
    }, 60 * 1000);
  }
});
</script>

<style scoped>
div#reading span {
  margin-left: 5px;
  margin-right: 1rem;
}

/* This ensures the reading icons aren't separated from the reading data point when wrapped on mobile */
@media screen and (min-width: 1000px) {
  div.readingData {
    display: inline;
  }
}
</style>
