<template>
  <v-container>
    <h2>System: {{ id }}</h2>
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
      system: {} as GardenSystem
    };
  },
  computed: {
    id() {
      return this.$route.params["id"];
    }
  },
  methods: {
    onMutation(mutation: MutationPayload) {
      if (
        mutation.type === "update" &&
        mutation.payload.Identifier === this.id
      ) {
        this.system = mutation.payload;
      }
    }
  },
  async created() {
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
