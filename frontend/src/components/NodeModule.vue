<template>
  <v-card class="card">
    <v-row class="parent-row">
      <v-col md="auto" :cols="auto">
      <v-icon class="node-icon">
        mdi-leaf
      </v-icon>
      </v-col>
      <v-col>
        <h1>{{ moduleName }}</h1>
          
        <p class="secondary-text">
          Id: 
          <router-link :to="'/graph/' + identifier">
            <code>{{ identifier }}</code>
          </router-link></p>
      </v-col>
      <v-col class="rhs" :align-items="center">
        <div v-if="isConnected">
          <h2 class="connected">
            Connected
            <v-icon class="icon">
              mdi-wifi-check
            </v-icon>
          </h2>
        </div>
        <div v-else>
          <h2 class="disconnected">
            Disconnected
            <v-icon class="icon">
              mdi-wifi-off
            </v-icon>
          </h2>
        </div>
        <p class="secondary-text">Last pushed {{ timeStamp }} seconds ago 
          <v-icon class="icon" small>
            mdi-clock
          </v-icon>
        </p>
      </v-col>
    </v-row>
  </v-card>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
    name: "NodeModule",
    props: ["moduleName", "identifier", "isConnected", "timestamp"]
});
</script>

<style scoped lang="scss">
  p a {
    padding-left: .25rem;
  }
  .card:hover {
    background: lighten($color: #1E1E1E, $amount: 10);
    transition: 0.25s;
  }
  .parent-row {
    padding: 0 2rem;
  }
  .secondary-text {
    color: $white-4;
  }
  .rhs {
    text-align: end;
  }
  .node-icon {
    font-size: 4.5em;
  }
  h2.connected {
    color: $green-0;  
    .icon {
      color: $green-0;  
    }
  }
  h2.disconnected {
    color: $red-0;  
    .icon {
      color: $red-0;  
    }
  }
  p .icon {
    padding-left: .25rem;
    color: $white-4;
  }
</style>
