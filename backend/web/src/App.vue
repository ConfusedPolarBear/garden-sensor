<template>
  <v-app>
    <v-app-bar app :color="getBarColor()" dark>
      <v-icon v-if="connected" id="logo" large>mdi-leaf</v-icon>
      <v-icon v-else-if="showConnectionError()" id="logo" large>
        mdi-leaf-off
      </v-icon>

      <h3 v-if="isMobile">Garden Management</h3>
      <h2 v-else>Garden Management</h2>

      <div id="wsIndicator" v-if="!connected && showConnectionError()">
        <h3 v-if="!isMobile" id="wsErrorText">Lost connection to server.</h3>

        <v-progress-circular
          :value="retryIn * 10"
          size="40"
          width="7"
          rotate="270"
        >
          {{ retryIn }}
        </v-progress-circular>
      </div>
    </v-app-bar>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script lang="ts">
import Vue from "vue";

let websocket: WebSocket;

export default Vue.extend({
  name: "App",
  data() {
    return {
      connected: false,
      retryIn: 10
    };
  },
  methods: {
    getBarColor(): string {
      const ok = "green darken-2";

      if (!this.showConnectionError()) {
        return ok;
      }

      return this.connected ? ok : "red";
    },
    showConnectionError(): boolean {
      // Don't show connection errors on the setup page
      return this.$router.currentRoute.path !== "/";
    },
    socketTryOpen() {
      const addr = window.localStorage.getItem("server");
      if (!addr) {
        return;
      }

      // Connect the websocket if it isn't connected
      if (websocket && this.connected) {
        return;
      }

      const ws = `${addr}/socket`.replace("http", "ws"); // protocol must be "ws" or "wss"
      console.debug(`[ws] opening websocket to ${ws}`);

      websocket = new WebSocket(ws);
      websocket.onopen = this.socketOpened;
      websocket.onerror = this.socketError;
      websocket.onclose = this.socketClosed;
      websocket.onmessage = this.socketMessage;
    },
    socketOpened() {
      console.debug("[ws] websocket opened");
      this.connected = true;
    },
    socketMessage(e: MessageEvent<string>) {
      const raw = JSON.parse(e.data);
      const type = raw.Type;
      const data = raw.Data;

      if (type === "register" || type === "update") {
        this.$store.commit(type, data);
      } else {
        console.warn(`[ws] unknown websocket message type ${type}`);
      }
    },
    socketError(e: Event) {
      console.error("[ws] websocket error", e);
      this.connected = false;
    },
    socketClosed() {
      console.debug("[ws] websocket closed");
      this.connected = false;
    }
  },
  computed: {
    isMobile(): boolean {
      return this.$vuetify.breakpoint.mobile;
    }
  },
  created() {
    this.socketTryOpen();

    // Reconnect the websocket at regular intervals if it is disconnected.
    // TODO: when the user first sets the address of the API server, the WS should be immediately setup as opposed to waiting
    setInterval(() => {
      if (this.connected) {
        this.retryIn = 10;
        return;
      }

      this.retryIn--;
      if (this.retryIn < 0) {
        this.retryIn = 10;
        this.socketTryOpen();
      }
    }, 1000);
  }
});
</script>

<style scoped>
#logo {
  margin-right: 0.5rem;
}

div#wsIndicator {
  margin-left: 1rem;
}

div#wsIndicator > * {
  margin-right: 1rem;
}

#wsErrorText {
  display: inline;
}
</style>
