<template>
  <v-card class="card" v-bind:class="{ selected: isSelected }" @click="clicked">
    <v-card-title>
      {{ names[type] }}
    </v-card-title>
    <v-card-text>
      {{ text[type] }}
    </v-card-text>
  </v-card>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "WizardCard",
  props: ["type", "selected"],
  data() {
    return {
      names: {
        wifi: "Wi-Fi",
        mesh: "Mesh"
      },
      text: {
        wifi: "Connect this system to my server through MQTT over Wi-Fi.",
        mesh: "Connect this system to the mesh network created by my current garden systems."
      }
    };
  },
  computed: {
    isSelected(): boolean {
      return this.type == this.selected;
    }
  },
  methods: {
    clicked() {
      this.$emit("click", this.$props.type);
    }
  }
});
</script>

<style scoped lang="scss">
$border: 2px solid;

.card {
  max-width: 300px;
  border: $border transparent;

  margin-right: 10px;
}

.card:hover {
  border: $border rgb(31, 147, 255);
  cursor: pointer;
}

.card.selected:not(:hover) {
  border: $border rgb(0, 96, 185);
}

.v-card--link:focus:before {
  // Disable the slight gray highlight that remains on a card while it is focused since it is the exact same color
  // as the stepper background.
  opacity: 0;
}
</style>
