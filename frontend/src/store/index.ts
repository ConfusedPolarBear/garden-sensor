import Vue from "vue";
import Vuex from "vuex";
import { GardenSystem } from "./types";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    systems: Array<GardenSystem>()
  },
  mutations: {
    // A garden system has just published new data.
    update(state, newClient) {
      for (let i = 0; i < state.systems.length; i++) {
        const c = state.systems[i];
        if (c.Identifier === newClient.Identifier) {
          Vue.set(state.systems, i, newClient);
          return;
        }
      }

      throw new Error(`Unable to find client with identifier ${newClient.Identifier}`);
    },
    // A new client has joined the server, add it's information to our local state
    register(state, systems) {
      state.systems = systems;
    }
  },
  actions: {},
  modules: {}
});
