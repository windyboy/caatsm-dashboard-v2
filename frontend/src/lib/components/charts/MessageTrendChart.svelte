<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    Chart,
    type ChartConfiguration,
    LineController,
    LineElement,
    PointElement,
    LinearScale,
    CategoryScale,
    Tooltip,
    Legend,
    Filler,
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

  // Generate static mock data once
  function generateMockData(): TrendDataPoint[] {
    const now = new Date();
    const generated: TrendDataPoint[] = [];
    for (let i = 23; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 60 * 1000);
      generated.push({
        time: time.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" }),
        count: Math.floor(Math.random() * 100) + 50,
      });
    }
    return generated;
  }

  // Simple chart data - no derived chains
  function getChartData() {
    const sourceData = data.length > 0 ? data : generateMockData();
    
    return {
      labels: sourceData.map((d) => d.time),
      datasets: [
        {
          label: "Messages",
          data: sourceData.map((d) => d.count),
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

  // Create chart once on mount
  onMount(() => {
    if (!canvasElement) return;

    const ctx = canvasElement.getContext("2d");
    if (!ctx) return;

    // Register required components (only once)
    Chart.register(
      LineController,
      LineElement,
      PointElement,
      LinearScale,
      CategoryScale,
      Tooltip,
      Legend,
      Filler
    );

    const config: ChartConfiguration<"line"> = {
      type: "line",
      data: getChartData(),
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

    // Only update if data prop changes (manual update, not reactive)
    return () => {
      // Cleanup handled in onDestroy
    };
  });

  // Chart updates are handled manually - no reactive effects to avoid loops

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
