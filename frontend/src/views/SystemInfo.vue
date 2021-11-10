<template>
  <v-container>
    <h2>System: {{ id }}</h2>
    <v-container>
      <v-row justify="space-around" justify-sm="start">
        <v-col>
          <!-- TODO: bring up individual sensor graphs -->
          <router-link :to="'/graph/' + id">
            <sensor-preview
              name="Temperature"
              icon="mdi-thermometer"
              :data="system.LastReading.Temperature"
            />
          </router-link>
        </v-col>
        <v-col>
          <router-link :to="'/graph/' + id">
            <sensor-preview
              name="Humidity"
              icon="mdi-water-percent"
              :data="system.LastReading.Temperature"
            />
          </router-link>
        </v-col>
      </v-row>
    </v-container>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import { MutationPayload } from "vuex";

import api, { GardenSystem } from "@/plugins/api";
import SensorPreview from "@/components/SensorPreview.vue";

export default Vue.extend({
  name: "SystemInfo",
  components: { SensorPreview },
  data() {
    return {
      reveal: false,
      system: {} as GardenSystem
    };
  },
  computed: {
    id() {
      return this.$route.params["id"];
    }
  },
  methods: {
    onMutation(mutation: MutationPayload, state: any) {
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
