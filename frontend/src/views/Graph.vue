<template>
  <v-container>
    <line-graph :id="id" :type="type" />
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import LineGraph from "@/components/LineGraph.vue";

export default Vue.extend({
  name: "Graph",
  components: { LineGraph },
  computed: {
    id() {
      return this.$route.params["id"];
    },
    type() {
      return this.$route.params["type"];
    }
  }
});
</script>
