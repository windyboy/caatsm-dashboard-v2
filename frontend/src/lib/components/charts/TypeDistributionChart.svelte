<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { Chart, type ChartConfiguration, ArcElement, Tooltip, Legend } from "chart.js";
  import Card from "../ui/Card.svelte";
  import { generateColors, defaultChartOptions } from "$lib/utils/chart-helpers";

  interface TypeDistributionChartProps {
    data?: Record<string, number>;
    title?: string;
  }

  let { data = {}, title = "Message Type Distribution" }: TypeDistributionChartProps = $props();

  let canvasElement: HTMLCanvasElement | null = $state(null);
  let chartInstance: Chart<"doughnut"> | null = $state(null);
  let updateTimer: ReturnType<typeof setTimeout> | null = $state(null);

  const chartData = $derived.by(() => {
    const entries = Object.entries(data);
    if (entries.length === 0) {
      return {
        labels: [],
        datasets: [],
      };
    }

    const labels = entries.map(([key]) => key);
    const values = entries.map(([, value]) => value);
    const colors = generateColors(labels.length);

    return {
      labels,
      datasets: [
        {
          label: "Messages",
          data: values,
          backgroundColor: colors,
          borderColor: "#fff",
          borderWidth: 2,
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
      if (!canvasElement || !chartData.labels.length) return;

      const ctx = canvasElement.getContext("2d");
      if (!ctx) return;

      // Update existing chart if it exists, otherwise create new one
      if (chartInstance) {
        chartInstance.data = chartData;
        chartInstance.update("none");
      } else {
        // Register required components
        Chart.register(ArcElement, Tooltip, Legend);

        const config: ChartConfiguration<"doughnut"> = {
          type: "doughnut",
          data: chartData,
          options: {
            ...defaultChartOptions,
            plugins: {
              ...defaultChartOptions.plugins,
              legend: {
                ...defaultChartOptions.plugins?.legend,
                position: "bottom" as const,
              },
            },
            cutout: "60%",
          },
        };

        chartInstance = new Chart(ctx, config);
      }
    }, 100);

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
    </div>
    {#if chartData.labels.length > 0}
      <div class="chart-wrapper">
        <canvas bind:this={canvasElement}></canvas>
      </div>
    {:else}
      <div class="chart-empty">
        <p class="muted">No data available</p>
      </div>
    {/if}
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

  .chart-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 300px;
    color: #71717a;
  }
</style>
