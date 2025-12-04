<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    Chart,
    type ChartConfiguration,
    LineElement,
    PointElement,
    LinearScale,
    CategoryScale,
    Tooltip,
    Legend,
  } from "chart.js";
  import Card from "../ui/Card.svelte";
  import { defaultChartOptions, chartColors } from "$lib/utils/chart-helpers";

  export interface TrendDataPoint {
    time: string;
    count: number;
  }

  interface MessageTrendChartProps {
    data?: TrendDataPoint[];
    title?: string;
  }

  let { data = [], title = "Message Trend" }: MessageTrendChartProps = $props();

  let canvasElement: HTMLCanvasElement | null = $state(null);
  let chartInstance: Chart<"line"> | null = $state(null);
  let updateTimer: ReturnType<typeof setTimeout> | null = $state(null);

  // Generate mock data if no data provided
  const chartData = $derived.by(() => {
    if (data.length > 0) {
      return {
        labels: data.map((d) => d.time),
        datasets: [
          {
            label: "Messages",
            data: data.map((d) => d.count),
            borderColor: chartColors.accent.blue,
            backgroundColor: `${chartColors.accent.blue}20`,
            tension: 0.4,
            fill: true,
            pointRadius: 3,
            pointHoverRadius: 5,
          },
        ],
      };
    }

    // Mock data for demonstration
    const now = new Date();
    const mockData: TrendDataPoint[] = [];
    for (let i = 23; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 60 * 1000);
      mockData.push({
        time: time.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" }),
        count: Math.floor(Math.random() * 100) + 50,
      });
    }

    return {
      labels: mockData.map((d) => d.time),
      datasets: [
        {
          label: "Messages",
          data: mockData.map((d) => d.count),
          borderColor: chartColors.accent.blue,
          backgroundColor: `${chartColors.accent.blue}20`,
          tension: 0.4,
          fill: true,
          pointRadius: 3,
          pointHoverRadius: 5,
        },
      ],
    };
  });

  $effect(() => {
    if (!canvasElement) return;

    // Debounce chart updates
    if (updateTimer) {
      clearTimeout(updateTimer);
    }

    updateTimer = setTimeout(() => {
      if (!canvasElement) return;

      const ctx = canvasElement.getContext("2d");
      if (!ctx) return;

      // Update existing chart if it exists, otherwise create new one
      if (chartInstance) {
        chartInstance.data = chartData;
        chartInstance.update("none");
      } else {
        // Register required components
        Chart.register(LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend);

        const config: ChartConfiguration<"line"> = {
          type: "line",
          data: chartData,
          options: {
            ...defaultChartOptions,
            plugins: {
              ...defaultChartOptions.plugins,
              legend: {
                ...defaultChartOptions.plugins?.legend,
                display: false,
              },
            },
            interaction: {
              intersect: false,
              mode: "index",
            },
          },
        };

        chartInstance = new Chart(ctx, config);
      }
    }, 150);

    return () => {
      if (updateTimer) {
        clearTimeout(updateTimer);
      }
      if (chartInstance) {
        chartInstance.destroy();
        chartInstance = null;
      }
    };
  });

  onDestroy(() => {
    if (chartInstance) {
      chartInstance.destroy();
      chartInstance = null;
    }
  });
</script>

<Card>
  <div class="chart-container">
    <div class="chart-header">
      <p class="eyebrow">{title}</p>
      <p class="muted">Messages over time (last 24 hours)</p>
    </div>
    <div class="chart-wrapper">
      <canvas bind:this={canvasElement}></canvas>
    </div>
  </div>
</Card>

<style>
  .chart-container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-height: 300px;
  }

  .chart-header {
    margin-bottom: 0.5rem;
  }

  .chart-wrapper {
    position: relative;
    height: 300px;
    width: 100%;
  }
</style>
