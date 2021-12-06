<template>
  <router-link :to="'/system/' + identifier" tag="v-card">
  <v-card v-ripple flat class="rounded-0 card">
    <v-row class="parent-row">
      <v-col md="auto" :cols="auto">
        <v-icon class="node-icon"> mdi-leaf </v-icon>
      </v-col>
      <v-col>
        <h1>{{ moduleName }}</h1>
        <p class="secondary-text">
          Id:
          <router-link :to="'/graph/' + identifier">
            <code>{{ identifier }}</code>
          </router-link>
        </p>
      </v-col>
      <v-col class="rhs" :align-items="center">
        <v-container v-if="isConnected">
            <h2 class="connected">
              <tooltip :text="meshInfo(announcement)">
                <span>
                  Connected
                  <v-icon class="icon"> mdi-wifi-check </v-icon>
                </span>
              </tooltip>
            </h2>
        </v-container>
        <v-container v-else>
            <h2 class="disconnected">
              <tooltip :text="meshInfo(announcement)">
                <span>
                  Disconnected
                  <v-icon class="icon"> mdi-wifi-off </v-icon>
                </span>
              </tooltip>
            </h2>
        </v-container>
        <p class="secondary-text">
          Last pushed {{ timestamp }} seconds ago
          <v-icon class="icon" small> mdi-clock </v-icon>
        </p>
      </v-col>
    </v-row>
  </v-card>
  </router-link>
</template>

<script lang="ts">
import Vue from "vue";
import { GardenSystem, GardenSystemInfo } from "@/store/types";
import Tooltip from "@/components/Tooltip.vue";


export default Vue.extend({
  name: "NodeModule",
  components: { Tooltip },
  props: ["moduleName", "identifier", "isConnected", "timestamp", "announcement"],
  methods: {
    meshInfo(announcement: GardenSystemInfo): any {
      const mesh = announcement.IsMesh;
      return mesh ? "Mesh" : `MQTT (CH ${announcement.Channel})`;
    },
  }
});
</script>

<style scoped lang="scss">
p a {
  padding-left: 0.25rem;
}
.card:hover {
  background: lighten($color: #1e1e1e, $amount: 8);
  transition: 0.25s;
}
.parent-row {
  padding: 0 2rem;
  margin: 0;
}
.secondary-text {
  color: $white-4;
}
.rhs {
  text-align: end;
}
.node-icon {
  font-size: 4.5em;
}
h2.connected {
  color: $green-0;
  .icon {
    color: $green-0;
  }
}
h2.disconnected {
  color: $red-0;
  .icon {
    color: $red-0;
  }
}
p .icon {
  padding-left: 0.25rem;
  color: $white-4;
}
</style>
