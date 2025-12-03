<!--
  @component MessageRateChart
  Displays time-series chart of message counts over time.
  
  Features:
  - Line chart showing message rate over time
  - Supports different time intervals (hour, day)
  - Updates based on time range selection
-->

<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { timeRange } from "$lib/stores/data/timeRange";
  import { createLogger } from "$lib/utils/logger";
  import { request } from "$lib/services/api";

  const logger = createLogger("MessageRateChart");

  let chartData = $state<{ time: string; count: number }[]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let chartContainer = $state<HTMLCanvasElement>();

  interface HistoricalStats {
    interval: string;
    data: Array<{ time: string; count: number }>;
  }

  async function loadChartData() {
    if (typeof window === "undefined") {
      return;
    }

    loading = true;
    error = null;

    try {
      const currentRange = $timeRange;
      const { start_time, end_time } = timeRange.getISOStrings(currentRange);

      // Determine interval based on time range
      const rangeMs = new Date(end_time).getTime() - new Date(start_time).getTime();
      const rangeDays = rangeMs / (1000 * 60 * 60 * 24);
      const interval = rangeDays > 7 ? "day" : "hour";

      const params = new URLSearchParams({
        start_time,
        end_time,
        interval,
      });

      const data = await request<HistoricalStats>(`/api/stats/historical?${params.toString()}`);
      chartData = data.data || [];
      logger.debug("Chart data loaded", { points: chartData.length, interval });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to load chart data";
      error = errorMessage;
      logger.error("Failed to load chart data", err);
    } finally {
      loading = false;
    }
  }

  function drawChart() {
    if (!chartContainer || chartData.length === 0) {
      return;
    }

    const canvas = chartContainer;
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return;
    }

    // Set canvas size
    const rect = canvas.getBoundingClientRect();
    const dpr = window.devicePixelRatio || 1;
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const height = rect.height;
    const padding = { top: 20, right: 20, bottom: 30, left: 50 };

    // Check theme once for the entire function
    const isDark = document.documentElement.classList.contains("dark");

    // Clear canvas
    ctx.clearRect(0, 0, width, height);

    // Get min/max values for scaling
    const counts = chartData.map((d) => d.count);
    const maxCount = Math.max(...counts, 1); // At least 1 to avoid division by zero
    const minCount = Math.min(...counts, 0);

    // Calculate chart area
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Draw background (transparent, let card background show through)
    ctx.clearRect(0, 0, width, height);

    // Draw grid lines
    ctx.strokeStyle = isDark ? "#334155" : "#e2e8f0"; // slate-700 or slate-200
    ctx.lineWidth = 1;

    // Horizontal grid lines (5 lines)
    for (let i = 0; i <= 5; i++) {
      const y = padding.top + (chartHeight * i) / 5;
      ctx.beginPath();
      ctx.moveTo(padding.left, y);
      ctx.lineTo(width - padding.right, y);
      ctx.stroke();
    }

    // Vertical grid lines (sparse, only at start and end)
    ctx.beginPath();
    ctx.moveTo(padding.left, padding.top);
    ctx.lineTo(padding.left, height - padding.bottom);
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(width - padding.right, padding.top);
    ctx.lineTo(width - padding.right, height - padding.bottom);
    ctx.stroke();

    // Draw line chart
    if (chartData.length > 0) {
      ctx.strokeStyle = "#3b82f6"; // brand-500
      ctx.lineWidth = 2;
      ctx.fillStyle = "rgba(59, 130, 246, 0.1)"; // brand-500 with opacity

      ctx.beginPath();
      const stepX = chartWidth / Math.max(chartData.length - 1, 1);

      chartData.forEach((point, index) => {
        const x = padding.left + index * stepX;
        const normalizedCount = maxCount > minCount 
          ? (point.count - minCount) / (maxCount - minCount)
          : 0.5;
        const y = padding.top + chartHeight * (1 - normalizedCount);

        if (index === 0) {
          ctx.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }
      });

      // Fill area under line
      ctx.lineTo(padding.left + (chartData.length - 1) * stepX, padding.top + chartHeight);
      ctx.lineTo(padding.left, padding.top + chartHeight);
      ctx.closePath();
      ctx.fill();

      // Draw line
      ctx.stroke();

      // Draw data points
      ctx.fillStyle = "#3b82f6";
      chartData.forEach((point, index) => {
        const x = padding.left + index * stepX;
        const normalizedCount = maxCount > minCount 
          ? (point.count - minCount) / (maxCount - minCount)
          : 0.5;
        const y = padding.top + chartHeight * (1 - normalizedCount);

        ctx.beginPath();
        ctx.arc(x, y, 3, 0, 2 * Math.PI);
        ctx.fill();
      });
    }

    // Draw Y-axis labels
    ctx.fillStyle = isDark ? "#94a3b8" : "#475569"; // slate-400 or slate-600
    ctx.font = "11px system-ui";
    ctx.textAlign = "right";
    ctx.textBaseline = "middle";

    for (let i = 0; i <= 5; i++) {
      const value = minCount + ((maxCount - minCount) * (5 - i)) / 5;
      const y = padding.top + (chartHeight * i) / 5;
      ctx.fillText(Math.round(value).toLocaleString(), padding.left - 10, y);
    }

    // Draw X-axis labels (first, middle, last)
    ctx.textAlign = "center";
    ctx.textBaseline = "top";

    if (chartData.length > 0) {
      const formatTime = (timeStr: string) => {
        const date = new Date(timeStr);
        const rangeMs = new Date(chartData[chartData.length - 1].time).getTime() - 
                       new Date(chartData[0].time).getTime();
        const rangeDays = rangeMs / (1000 * 60 * 60 * 24);
        
        if (rangeDays > 7) {
          return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
        } else {
          return date.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit" });
        }
      };

      // First label
      const firstX = padding.left;
      ctx.fillText(formatTime(chartData[0].time), firstX, height - padding.bottom + 5);

      // Last label
      const lastX = padding.left + (chartData.length - 1) * (chartWidth / Math.max(chartData.length - 1, 1));
      ctx.fillText(formatTime(chartData[chartData.length - 1].time), lastX, height - padding.bottom + 5);

      // Middle label (if more than 2 points)
      if (chartData.length > 2) {
        const midIndex = Math.floor(chartData.length / 2);
        const midX = padding.left + midIndex * (chartWidth / Math.max(chartData.length - 1, 1));
        ctx.fillText(formatTime(chartData[midIndex].time), midX, height - padding.bottom + 5);
      }
    }
  }

  let unsubscribe: (() => void) | null = null;

  onMount(() => {
    // Subscribe to time range changes
    unsubscribe = timeRange.subscribe(() => {
      loadChartData();
    });

    // Redraw on window resize
    const handleResize = () => {
      if (chartData.length > 0) {
        drawChart();
      }
    };
    window.addEventListener("resize", handleResize);

    return () => {
      if (unsubscribe) unsubscribe();
      window.removeEventListener("resize", handleResize);
    };
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  // Redraw when chartData or chartContainer changes
  $effect(() => {
    if (chartData.length > 0 && chartContainer && !loading) {
      // Use requestAnimationFrame to ensure DOM is ready
      requestAnimationFrame(() => {
        drawChart();
      });
    }
  });
</script>

<div class="chart-card" role="region" aria-label="Message Rate Chart">
  <div class="chart-header">
    <div class="chart-icon">
      <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
        ></path>
      </svg>
    </div>
    <p class="chart-title">Message Rate Over Time</p>
  </div>

  <div class="chart-content">
    {#if loading}
      <div class="chart-loading">
        <p class="text-sm text-slate-500">Loading chart data...</p>
      </div>
    {:else if error}
      <div class="chart-error">
        <p class="text-sm font-semibold text-red-600 mb-1">Error</p>
        <p class="text-xs text-slate-500">{error}</p>
      </div>
    {:else if chartData.length === 0}
      <div class="chart-empty">
        <p class="text-sm text-slate-500">No data available for the selected time range</p>
      </div>
    {:else}
      <div class="chart-wrapper">
        <canvas bind:this={chartContainer} class="chart-canvas"></canvas>
        <!-- Simple text-based visualization as placeholder -->
        <div class="chart-stats">
          <p class="text-xs text-slate-400">
            {chartData.length} data points | 
            Max: {Math.max(...chartData.map(d => d.count))} messages | 
            Total: {chartData.reduce((sum, d) => sum + d.count, 0).toLocaleString()} messages
          </p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .chart-card {
    border-radius: 0.5rem;
    padding: 1.25rem;
    transition:
      transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1),
      background-color 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    transform: scale(1);
    background: var(--stats-card-bg);
    box-shadow: var(--stats-card-shadow);
    border: var(--stats-card-border);
  }

  .chart-card:hover {
    background: var(--stats-card-hover-bg);
    box-shadow: var(--stats-card-hover-shadow);
    transform: scale(1.05);
  }

  .chart-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .chart-icon {
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(191, 219, 254, 0.5);
    background: linear-gradient(to right, #dbeafe, #f3e8ff);
  }

  .chart-title {
    font-size: 0.75rem;
    line-height: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-weight: 700;
    background: linear-gradient(to right, #2563eb, #9333ea);
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .chart-content {
    min-height: 200px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .chart-loading,
  .chart-error,
  .chart-empty {
    text-align: center;
    padding: 2rem 0;
  }

  .chart-wrapper {
    width: 100%;
    height: 200px;
    position: relative;
  }

  .chart-canvas {
    width: 100%;
    height: 100%;
  }

  .chart-stats {
    margin-top: 0.5rem;
    text-align: center;
  }
</style>
