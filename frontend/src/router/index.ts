import Vue from "vue";
import VueRouter, { RouteConfig } from "vue-router";
import Setup from "@/views/Setup.vue";
import Systems from "@/views/Systems.vue";
import Graph from "@/views/Graph.vue";
import SystemWizard from "@/views/SystemWizard.vue";

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
  },
  {
    path: "/graph/:id",
    name: "graph",
    component: Graph
  },
  {
    path: "/configure",
    name: "configure",
    component: SystemWizard
  }
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes
});

export default router;
