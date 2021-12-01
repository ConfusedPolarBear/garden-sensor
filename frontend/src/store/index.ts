import Vue from "vue";
import Vuex from "vuex";
import { GardenSystem, StoreState } from "./types";

Vue.use(Vuex);

export default new Vuex.Store<StoreState>({
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

      // If we get here, it's because this is a new system, append it to the end of the array.
      state.systems.push(newClient);
    },
    // A new client has joined the server, add it's information to our local state
    register(state, systems) {
      state.systems = systems;
    }
  },
  actions: {},
  modules: {}
});
