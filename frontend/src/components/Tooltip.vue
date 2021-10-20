<template>
  <span>
    <v-tooltip bottom color="black">
      <template v-slot:activator="{ on, attrs }">
        <span v-on="on" v-bind="attrs">
          <slot />
        </span>
      </template>

      <strong>{{ text }}</strong>
    </v-tooltip>
  </span>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "Tooltip",
  props: ["text"]
});
</script>
