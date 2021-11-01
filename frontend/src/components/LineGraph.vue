<template>
  <v-container>
    <div v-show="valid">
      <canvas :id="type + '-graph'" width="500px" height="300px" />
      <br />

      <table>
        <thead>
          <td>Maximum</td>
          <td>Minimum</td>
          <td>Average</td>
          <td>Discarded</td>
          <td>Total points</td>
        </thead>
        <tbody>
          <td>{{ max }}</td>
          <td>{{ min }}</td>
          <td>{{ avg }}</td>
          <td>{{ errors }}</td>
          <td>{{ points.length }}</td>
        </tbody>
      </table>
    </div>

    <div v-show="!valid">
      <v-progress-circular v-if="loading" indeterminate />
      <v-alert type="error" v-else>
        <strong>
          <span>No valid {{ type.toLowerCase() }} data points</span>
        </strong>
      </v-alert>
    </div>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import api from "@/plugins/api";
import { MutationPayload } from "vuex";

import {
  Chart,
  ArcElement,
  LineElement,
  BarElement,
  PointElement,
  BarController,
  BubbleController,
  DoughnutController,
  LineController,
  PieController,
  PolarAreaController,
  RadarController,
  ScatterController,
  CategoryScale,
  LinearScale,
  LogarithmicScale,
  RadialLinearScale,
  TimeScale,
  TimeSeriesScale,
  Decimation,
  Filler,
  Legend,
  Title,
  Tooltip,
  SubTitle,
  ChartData,
  ChartConfiguration
} from "chart.js";

export default Vue.extend({
  name: "LineGraph",
  props: ["id", "type"],
  data() {
    return {
      system: {} as any,
      chart: {} as Chart,

      loading: true,
      labels: Array<string>(),
      points: Array<number>(),

      // Numerical day of the previous point. Used to print a short date when moving from one day to the next.
      lastPoint: 99,

      errors: 0,
      min: 0.0,
      max: 0.0,
      avg: 0.0
    };
  },
  methods: {
    onMutation(mutation: MutationPayload, state: any) {
      if (mutation.type !== "update") {
        return;
      }

      const payload = mutation.payload;

      // Ensure that the update references the current system
      if (this.$route.path.indexOf(payload.Identifier) === -1) {
        return;
      }

      const p = payload.LastReading;

      this.storePoint(p);
      this.update();
    },
    storePoint(point: any) {
      const p = point[this.type];

      if (point.Error || p == 32768) {
        this.errors++;
        return;
      }

      this.labels.push(this.parseDate(point.CreatedAt));
      this.points.push(p);
    },
    update(): void {
      this.loading = false;

      this.chart.update();

      let n = this.points.length;
      this.min = Number.MAX_SAFE_INTEGER;
      this.max = Number.MIN_SAFE_INTEGER;

      for (const p of this.points) {
        if (p < this.min) {
          this.min = p;
        }

        if (p > this.max) {
          this.max = p;
        }

        this.avg += p;
      }

      this.avg /= n;
      this.avg = Number(this.avg.toFixed(2));
    },
    parseDate(raw: Date): string {
      const opts = { hour12: false } as Intl.DateTimeFormatOptions;
      opts.timeStyle = "medium";

      const d = new Date(raw);
      let label = d.toLocaleTimeString("default", opts).replace(/^24:/, "00:"); // convert the time "24:30:24" into "00:30:24"

      // Visually flag when the next day happens
      if (d.getDate() > this.lastPoint) {
        const month = d.toLocaleDateString("default", { month: "long" });
        label = `${month} ${d.getDate()}`;
      }

      this.lastPoint = d.getDate();

      return label;
    }
  },
  computed: {
    valid() {
      return this.$data.points.length > 0;
    },
    color() {
      switch (this.type) {
        case "Temperature":
          return "lightblue";

        case "Humidity":
          return "lightgreen";

        default:
          return "red";
      }
    }
  },
  async created() {
    const res = await api(`/system/${this.$route.params.id}`);
    this.system = await res.json();

    // The backend sends readings ordered newest to oldest but since we store points with push(), the graph will be mirrored.
    const readings = this.system.Readings.reverse();

    for (const raw of readings) {
      this.storePoint(raw);
    }

    const temperatureCanvas = document.querySelector(
      `#${this.type}-graph`
    ) as HTMLCanvasElement;
    if (!temperatureCanvas) {
      console.error("unable to find graph canvas");
      return;
    }

    const data = {
      labels: this.labels,
      datasets: [
        {
          label: this.type,
          data: this.points,
          backgroundColor: this.color,
          borderColor: this.color
        }
      ]
    } as ChartData;

    const config = {
      type: "line",
      data: data,
      options: {
        scales: {
          y: {
            suggestedMin: 0,
            grid: {
              color: "gray"
            }
          }
        }
      }
    } as ChartConfiguration;

    this.chart = new Chart(temperatureCanvas, config);

    this.update();
  },
  mounted() {
    Chart.register(
      ArcElement,
      LineElement,
      BarElement,
      PointElement,
      BarController,
      BubbleController,
      DoughnutController,
      LineController,
      PieController,
      PolarAreaController,
      RadarController,
      ScatterController,
      CategoryScale,
      LinearScale,
      LogarithmicScale,
      RadialLinearScale,
      TimeScale,
      TimeSeriesScale,
      Decimation,
      Filler,
      Legend,
      Title,
      Tooltip,
      SubTitle
    );

    this.$store.subscribe(this.onMutation);
  }
});
</script>

<style scoped>
table > thead td {
  width: 100px;
  font-weight: bolder;
}
</style>
