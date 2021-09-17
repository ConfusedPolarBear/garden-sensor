<template>
  <v-container>
    <p>Systems:</p>
    <v-data-table :items="$store.state.systems" :headers="headers">
      <template v-slot:[`item.Identifier`]="{ item }">
        <code>{{ item.Identifier }}</code>
      </template>

      <template v-slot:[`item.LastReading`]="{ item }">
        <div id="reading" v-if="dataValid(item.LastSeen)">
          <v-icon>mdi-thermometer</v-icon>
          <span>{{ item.LastReading.Temperature }} °C</span>

          <v-icon>mdi-water-percent</v-icon>
          <span>{{ item.LastReading.Humidity }} %</span>

          <v-icon>mdi-clock</v-icon>
          <span>{{ age(item.LastSeen) }}</span>
        </div>
        <div id="reading" v-else>
          <v-icon>mdi-clock</v-icon>
          <span>has not reported data yet</span>
        </div>
      </template>
    </v-data-table>
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
      systems: Array<unknown>()
    };
  },
  methods: {
    load(): void {
      // Load all initial systems.
      api("/systems")

      // Subscribe to the Vuex store for future updates.
      this.$store.subscribe(this.onMutation);
    },
    onMutation(mutation: MutationPayload, state: any) {
      if (mutation.type !== "register" && mutation.type !== "update") {
        return;
      }

      this.systems = state.systems;
    },
    dataValid(lastSeen: string): boolean {
      return !lastSeen.startsWith("0001");
    },
    age(lastSeen: string): string {
      let diff = Number(new Date()) - Number(new Date(lastSeen));
      diff = Number(diff) / 1000;
      return `last seen ${diff.toFixed(0)} seconds ago`;
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
</style>
