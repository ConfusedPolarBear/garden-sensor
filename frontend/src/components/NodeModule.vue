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
            <tooltip :text="lastSeen">
              Last pushed {{ humanize(timestamp) }}
              <v-icon class="icon" small> mdi-clock </v-icon>
            </tooltip>
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
  data() {
    return {
      dateOptions: {
        hour12: false
      } as Intl.DateTimeFormatOptions,

      // Maximum number of humanized duration portions to display. Higher values will be more accurate but
      // take up more space. For example, if a node was last seen 4 weeks ago, do most users really care
      // that it was *exactly* 4 weeks, 11 days, 13 hours, 7 minutes, 2 seconds ago?
      maximumParts: 2
    };
  },
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
    },
    humanize(raw: number): string {
      /* Example: suppose raw is 436507 (seconds).
       * Step 1: floor(436507 / 86400) = 5 days.
       * raw -= 5*86400. raw is now 4507
       * Step 2: floor(4507 / 3600) = 1 hour.
       * raw -= 1*3600. raw is now 907
       * Step 3: floor(907 / 60) = 15 minutes.
       * raw -= 15*60. raw is now 7 seconds.
       *
       * For all pushed parts, check if they need to be pluralized (almost certainly since only 1 doesn't need it).
       * Combine those parts into the string "5 days, 1 hour, 15 minutes, 7 seconds".
       * This is correct since 7 + 15*60 + 1*60*60 + 5*60*60*24 = 436507, which is exactly equal to the input value.
       */

      if (raw <= 2) {
        return "now";
      }

      interface duration {
        n: number;
        s: string;
      }

      let parts: Array<duration> = [];

      const steps = [
        {
          n: 60 * 60 * 24,
          s: "day"
        },
        {
          n: 60 * 60,
          s: "hour"
        },
        {
          n: 60,
          s: "minute"
        },
        {
          n: 1,
          s: "second"
        }
      ] as Array<duration>;

      // Convert the seconds into a human readable duration.
      for (const step of steps) {
        if (parts.length >= this.maximumParts) {
          break;
        }

        const tmp = Math.floor(raw / step.n);
        if (tmp == 0) {
          continue;
        }

        parts.push({ n: tmp, s: step.s });
        raw -= tmp * step.n;
      }

      // Fix grammar.
      let humanized = "";
      for (let i = 0; i < parts.length; i++) {
        const p = parts[i];
        if (p.n > 1) {
          p.s += "s";
        }

        humanized += `${p.n} ${p.s}, `;
      }

      // Remove trailing comma.
      return humanized.substring(0, humanized.length - 2) + " ago";
    }
  },
  computed: {
    state(): string {
      return this.isConnected ? "connected" : "disconnected";
    },
    lastSeen(): string {
      const t = this.$props.timestamp * 1000; // seconds -> ms
      return new Date(new Date().getTime() - t).toLocaleString(
        undefined,
        this.dateOptions
      );
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
