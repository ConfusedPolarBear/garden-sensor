<template>
  <v-card flat class="rounded-0 parent">
    <v-row no-gutters>
      <v-col md="auto" :cols="auto">
        <v-icon class="icon">{{ icon }}</v-icon>
      </v-col>
      <v-col>
        <h1 class="name">{{ name }}</h1>
        <h1 class="data">{{ data }}</h1>
      </v-col>
      <v-col class="rhs">
        <div>Id: {{ identifier }}</div>
        <div>{{ model }}</div>
      </v-col>
    </v-row>
  </v-card>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
    name: "SensorPreview",
    props: ["icon", "name", "data", "identifier", "model"]
});
</script>

<style scoped lang="scss">
  .parent {
    padding: 1rem;
    .name,h1 {
      padding-left: 1rem;
    }
  }
  .icon {
    padding-left: 1rem;
    text-align: left;
    font-size: 5em;
    color: $green-0;
  }  
  .name {
    font-weight: 300;
  }
  .data {
    line-height: 24px;
  }
  .rhs.col {
    color: $white-3;
    padding-right: 2rem;
  }
  .rhs {
    text-align: end;
    div {
      font-size: 18px;
    }
  }

</style>