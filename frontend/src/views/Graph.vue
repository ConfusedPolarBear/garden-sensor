<template>
  <v-container>
    <line-graph :id="id" type="Temperature" />

    <line-graph :id="id" type="Humidity" />
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import LineGraph from "@/components/LineGraph.vue";

export default Vue.extend({
  name: "System",
  components: { LineGraph },
  computed: {
    id() {
      return this.$route.params["id"];
    }
  }
});
</script>
