<template>
  <v-container>
    <p>
      Welcome to Garden Management. Before you can get started, we need to know
      the address of your garden backend server.
    </p>
    <v-form @submit.prevent="validateServer">
      <v-text-field v-model="server" label="API server" type="url" />
      <v-btn @click="validateServer" color="primary">Finish</v-btn>
    </v-form>

    <v-snackbar v-model="snackbar.show" color="red" timeout="3000">
      <strong>{{ snackbar.text }}</strong>
    </v-snackbar>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "Setup",
  data() {
    return {
      server: window.localStorage.getItem("server") || window.location.origin,
      snackbar: {
        show: false,
        text: ""
      }
    };
  },
  methods: {
    validateServer(): void {
      let server = this.server;
      if (!server.startsWith("http")) {
        this.showSnackbar("Address must start with a protocol");
        return;
      }

      if (server.endsWith("/")) {
        server = server.substr(0, server.length - 1);
      }
      server += "/ping";

      fetch(server).then(async (r) => {
        console.debug(r);

        if (r.status != 204) {
          // Test if the user entered the address of the web UI
          const text = await r.text();
          if (text.indexOf("<!-- garden web interface -->")) {
            this.showSnackbar(
              "Enter the address of the API server, not the web interface"
            );
          } else {
            this.showSnackbar("Invalid response from server");
          }
          return;
        }

        window.localStorage.setItem("server", this.server);
        this.$router.push("/systems");
      });
    },
    showSnackbar(text: string) {
      this.snackbar = {
        show: true,
        text: text
      };
    }
  },
  created() {
    this.validateServer();
  }
});
</script>
