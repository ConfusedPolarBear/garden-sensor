<template>
  <v-container>
    <p>Systems:</p>
    <div class="module-list-container" v-for="sys in $store.state.systems" :key="sys.Identifier">
      <node-module
          moduleName="Node Module"
          :identifier="sys.Identifier"
          :isConnected="true"
          :timestamp="age(sys.UpdatedAt)"
          :announcement="sys.Announcement"
        />
    </div>
    
    <br/>

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
import NodeModule from "@/components/NodeModule.vue";

export default Vue.extend({
  name: "Systems",
  components: { CommandDialog, NodeModule },
  // components: { CommandDialog },
  // state
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
  // actions
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
      return `${diff.toFixed(0)}`;
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
.module-list-container {
  overflow: hidden;
  border-radius: 4px;
}

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
