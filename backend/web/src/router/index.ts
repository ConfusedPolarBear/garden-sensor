import Vue from "vue";
import VueRouter, { RouteConfig } from "vue-router";
import Setup from "@/views/Setup.vue";
import Systems from "@/views/Systems.vue";

Vue.use(VueRouter);

const routes: Array<RouteConfig> = [
  {
    path: "/",
    name: "setup",
    component: Setup
  },
  {
    path: "/systems",
    name: "systems",
    component: Systems
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

export default router;
