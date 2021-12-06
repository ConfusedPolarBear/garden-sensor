<template>
  <v-row justify="center">
    <v-dialog v-model="rShow" width="700px" persistent>
      <v-card>
        <v-card-title>{{ title }}</v-card-title>

        <v-divider />

        <v-form id="command">
          <v-select
            v-model="predefined.command"
            @change="predefinedSelected"
            :items="predefined.available"
            label="Predefined commands"
          />

          <v-checkbox
            v-model="danger.checked"
            v-if="danger.dangerous"
            label="I understand that this is irreversible"
          />

          <v-expansion-panels v-model="panel">
            <v-expansion-panel v-model="panel">
              <v-expansion-panel-header>
                Send raw command
              </v-expansion-panel-header>

              <v-expansion-panel-content>
                <v-text-field
                  :rules="validateCommand()"
                  id="txtCommand"
                  v-model="command"
                  :counter="maximum"
                  label="Command"
                />
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-form>

        <v-card-actions>
          <v-btn color="secondary" @click="close(false)">Cancel</v-btn>

          <v-spacer />

          <v-btn color="primary" :disabled="disallowSend()" @click="close()">
            Send
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-row>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "CommandDialog",
  props: ["show", "id"],
  data() {
    return {
      predefined: {
        available: [
          { header: "Power" },
          {
            text: "Restart (including controller)",
            value: "Restart"
          },
          {
            text: "Restart (except controller)",
            value: "RestartEC"
          },
          { divider: true },
          { header: "Networking" },
          {
            text: "List local mesh peers",
            value: "ListPeers"
          },
          {
            text: "Ping",
            value: "Ping"
          },
          { divider: true },
          { header: "Danger zone" },
          {
            text: "Factory reset",
            value: "Reset"
          },
          { divider: true },
          { header: "Advanced" },
          {
            text: "Recalculate peer RSSI",
            value: "Scan"
          },
          {
            text: "Send raw command",
            value: ""
          }
        ],
        command: "Restart"
      },
      danger: {
        dangerous: false,
        checked: false
      },
      panel: -1,
      maximum: 210,
      command: ""
    };
  },
  computed: {
    rShow() {
      // computed property is used here to avoid the dialog popping up repeatedly.
      return this.show;
    },
    title(): string {
      // TODO: display friendly name and MAC here
      let id: string = this.id;

      if (id === "FFFFFFFFFFFF") {
        return "Broadcast command";
      }

      let colons = "";
      while (id.length > 0) {
        colons += `${id.substring(0, 2)}:`;
        id = id.substring(2, id.length);
      }

      colons = colons.substring(0, colons.length - 1);

      return `Send command to ${colons}`;
    }
  },
  methods: {
    validateCommand() {
      if (this.command === "") {
        return [false];
      }

      // Mesh commands are formatted like this: {'D':'dst-BEFF4D1AAD9D','Command':'COMMAND'}. Minimum overhead: 37 bytes.
      if (this.command.length > this.maximum) {
        return ["Command is too long"];
      }

      try {
        JSON.parse(this.command);
      } catch {
        return ["Command must be valid JSON"];
      }

      return [true];
    },
    disallowSend(): boolean {
      console.debug("checking if send is disallowed");

      if (this.validateCommand()[0] !== true) {
        return true;
      }

      return this.danger.dangerous && !this.danger.checked;
    },
    predefinedSelected() {
      const c = this.predefined.command;

      if (c === "RestartEC") {
        // Restart (except controller). This works because the sleep message must tell controllers to opt-in.
        this.command = `{"Command":"Sleep","Period":1}`;
        return;
      }

      if (c === "") {
        this.panel = 0;
      }

      this.command = `{"Command":"${c}"}`;
      this.danger.dangerous = c === "Reset";
    },
    close(emit = true): void {
      if (emit) {
        this.$emit("command", { id: this.id, command: this.command });
      }

      this.danger = {
        dangerous: false,
        checked: false
      };

      this.$emit("close");
    }
  },
  created() {
    this.predefinedSelected();
  }
});
</script>

<style>
form#command {
  margin: 0 15px;
  padding: 5px 0;
}

#txtCommand {
  font-family: monospace !important;
}
</style>
