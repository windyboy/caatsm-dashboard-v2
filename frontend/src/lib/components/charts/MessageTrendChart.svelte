<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { Chart, registerables, type ChartConfiguration } from "chart.js";
  import Card from "../ui/Card.svelte";
  import { defaultChartOptions, chartColors } from "$lib/utils/chart-helpers";
  import { _ } from "svelte-i18n";

  export interface TrendDataPoint {
    time: string;
    count: number;
  }

  interface MessageTrendChartProps {
    data?: TrendDataPoint[];
    title?: string;
  }

  let { data = [], title }: MessageTrendChartProps = $props();

  const chartTitle = $derived(title ?? $_("charts.messageTrend"));

  let canvasElement: HTMLCanvasElement | null = $state(null);
  let chartInstance: Chart<"line"> | null = $state(null);

  // Generate static mock data once
  function generateMockData(): TrendDataPoint[] {
    const now = new Date();
    const generated: TrendDataPoint[] = [];
    for (let i = 23; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 60 * 1000);
      generated.push({
        time: time.toISOString(), // Use ISO format like backend
        count: Math.floor(Math.random() * 100) + 50,
      });
    }
    return generated;
  }

  // Format time string for display
  function formatTime(timeStr: string): string {
    // If it's already a formatted time string (like "02:30 PM" or "14:30"), return as-is
    if (!timeStr.includes("T") && !timeStr.includes("Z") && timeStr.match(/^\d{1,2}:\d{2}/)) {
      return timeStr;
    }

    try {
      const date = new Date(timeStr);
      if (isNaN(date.getTime())) {
        return timeStr; // Return original string if parsing fails
      }
      return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" });
    } catch {
      return timeStr;
    }
  }

  // Memoized chart data using $derived for performance optimization
  const chartData = $derived.by(() => {
    const sourceData = data.length > 0 ? data : generateMockData();

    return {
      labels: sourceData.map((d) => formatTime(d.time)),
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
  });

  // Update chart data when data prop changes
  function updateChart() {
    if (!chartInstance) return;

    chartInstance.data.labels = chartData.labels;
    chartInstance.data.datasets[0].data = chartData.datasets[0].data;
    chartInstance.update();
  }

  // Create chart once on mount
  onMount(() => {
    if (!canvasElement) return;

    const ctx = canvasElement.getContext("2d");
    if (!ctx) return;

    // Register required components (only once)
    Chart.register(...registerables);

    const config: ChartConfiguration<"line"> = {
      type: "line",
      data: chartData,
      options: {
        ...(defaultChartOptions as Partial<ChartConfiguration<"line">["options"]>),
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

    return () => {
      // Cleanup handled in onDestroy
    };
  });

  // Reactively update chart when data changes
  $effect(() => {
    if (chartInstance) {
      updateChart();
    }
  });

  onDestroy(() => {
    if (chartInstance) {
      chartInstance.destroy();
      chartInstance = null;
    }
  });
</script>

<svelte:boundary>
  <Card>
    <div class="chart-container">
      <div class="chart-header">
        <p class="eyebrow">{chartTitle}</p>
        <p class="muted">{$_("dashboard.messages")} over time (last 24 hours)</p>
      </div>
      <div class="chart-wrapper">
        <canvas
          bind:this={canvasElement}
          aria-label="Message trend chart showing message count over the last 24 hours"
        ></canvas>
      </div>
    </div>
  </Card>
</svelte:boundary>

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
