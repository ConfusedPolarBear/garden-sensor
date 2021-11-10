<template>
  <v-card v-ripple class="parent">
    <v-row no-gutters><v-icon class="icon">{{ icon }}</v-icon></v-row>
    <v-row no-gutters><div class="name">{{ name }}</div></v-row>
    <v-row no-gutters><h1>{{ data }}</h1></v-row>
  </v-card>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
    name: "SensorPreview",
    props: ["icon", "name", "data"]
});
</script>

<style scoped lang="scss">
  .parent {
    width: 12em;
    height: 12em;
    .name,h1 {
      padding-left: 1rem;
    }
  }
  .icon {
    text-align: left;
    font-size: 4em;
    color: $green-0;
  }  
  .name {
    font-size: 20px;
    color: $white-3;
  }
  h1 {
    font-size: 3rem;
  }

</style>