<template>
  <router-link :to="'/system/' + identifier" custom v-slot="{ navigate }">
    <v-card
      @click="navigate"
      @keypress.enter="navigate"
      v-ripple
      flat
      class="rounded-0 card"
    >
      <v-row class="parent-row">
        <v-col md="auto" cols="auto">
          <v-icon class="node-icon"> mdi-leaf </v-icon>
        </v-col>
        <v-col>
          <h1>{{ moduleName }}</h1>
          <p class="secondary-text">
            <code>{{ identifier }}</code>
          </p>
        </v-col>
        <v-col class="rhs" align-items="center">
          <span>
            <h2 :class="state">
              <tooltip :text="meshInfo(announcement)">
                <span class="moduleState">
                  {{ state }}
                  <v-icon class="icon">
                    {{ isConnected ? "mdi-wifi-check" : "mdi-wifi-off" }}
                  </v-icon>
                </span>
              </tooltip>
            </h2>
          </span>
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
import { GardenSystemInfo } from "@/store/types";
import Tooltip from "@/components/Tooltip.vue";

export default Vue.extend({
  name: "NodeModule",
  components: { Tooltip },
  props: [
    "moduleName",
    "identifier",
    "isConnected",
    "timestamp",
    "announcement"
  ],
  methods: {
    meshInfo(announcement: GardenSystemInfo): string {
      const mesh = announcement.IsMesh;
      return mesh ? "Mesh" : `MQTT (CH ${announcement.Channel})`;
    }
  },
  computed: {
    state(): string {
      return this.isConnected ? "connected" : "disconnected";
    }
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

span.moduleState {
  text-transform: capitalize;
}

p .icon {
  padding-left: 0.25rem;
  color: $white-4;
}
</style>
