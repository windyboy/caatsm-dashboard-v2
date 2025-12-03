<script lang="ts">
  import { onMount } from "svelte";
  import LiveStream from "$lib/components/LiveStream.svelte";
  import StatsCard from "$lib/components/StatsCard.svelte";
  import HealthStatus from "$lib/components/HealthStatus.svelte";
  import TimeRangeSelector from "$lib/components/TimeRangeSelector.svelte";
  import MetricsCard from "$lib/components/MetricsCard.svelte";
  import WebSocketMetrics from "$lib/components/WebSocketMetrics.svelte";
  import MessageRateChart from "$lib/components/MessageRateChart.svelte";
  import ErrorBoundary from "$lib/components/ErrorBoundary.svelte";
  import { getStats } from "$lib/services/api";
  import { stats } from "$lib/stores/data/stats";
  import { timeRange } from "$lib/stores/data/timeRange";
  import { createLogger } from "$lib/utils/logger";

  const logger = createLogger("Dashboard");

  // Load statistics function
  async function loadStats() {
    // Only run in browser environment
    if (typeof window === "undefined") {
      return;
    }

    try {
      const currentRange = $timeRange;
      const { start_time, end_time } = timeRange.getISOStrings(currentRange);
      
      logger.debug("Loading statistics from server", {
        start_time,
        end_time,
        preset: currentRange.preset,
      });
      
      const serverStats = await getStats(start_time, end_time);
      
      // Validate and normalize response structure
      if (!serverStats || typeof serverStats !== 'object') {
        logger.warn("Invalid statistics response", { serverStats });
        return;
      }
      
      // Ensure all required fields exist with defaults
      const validatedStats = {
        total: serverStats.total ?? 0,
        byPriority: serverStats.byPriority ?? {},
        byType: serverStats.byType ?? {},
      };
      
      stats.setStats(validatedStats);
      logger.info("Statistics loaded successfully", {
        total: validatedStats.total,
        byPriorityCount: Object.keys(validatedStats.byPriority).length,
        byTypeCount: Object.keys(validatedStats.byType).length,
        preset: currentRange.preset,
      });
    } catch (error) {
      logger.error("Failed to load statistics", error);
      // Don't throw - let WebSocket updates handle it
    }
  }

  // Load statistics on page mount and when time range changes
  onMount(() => {
    let isInitialLoad = true;
    
    // Subscribe to time range changes
    const unsubscribe = timeRange.subscribe(() => {
      // Skip the initial callback to avoid duplicate API call
      if (isInitialLoad) {
        isInitialLoad = false;
        return;
      }
      loadStats();
    });
    
    // Load initial stats after subscription is set up
    loadStats();
    
    return () => {
      unsubscribe();
    };
  });
</script>

<main class="mx-auto max-w-7xl px-4 sm:px-6 py-8 sm:py-10">
  <section class="space-y-10">
    <!-- Time Range Selector -->
    <div class="w-full">
      <ErrorBoundary context={{ component: "TimeRangeSelector" }}>
        <TimeRangeSelector />
      </ErrorBoundary>
    </div>

    <!-- Main content: Live Stream and Stats Cards -->
    <div class="grid gap-8 lg:grid-cols-3">
      <!-- Live Stream - Full width on mobile, 2/3 on desktop -->
      <div class="lg:col-span-2">
        <ErrorBoundary context={{ component: "LiveStream" }}>
          <LiveStream />
        </ErrorBoundary>
      </div>

      <!-- Stats Cards - Stack on mobile, sidebar on desktop -->
      <div class="space-y-6">
        <div>
          <ErrorBoundary context={{ component: "HealthStatus" }}>
            <HealthStatus />
          </ErrorBoundary>
        </div>
        <div>
          <ErrorBoundary context={{ component: "MetricsCard" }}>
            <MetricsCard />
          </ErrorBoundary>
        </div>
        <div>
          <ErrorBoundary context={{ component: "WebSocketMetrics" }}>
            <WebSocketMetrics />
          </ErrorBoundary>
        </div>
        <div>
          <ErrorBoundary context={{ component: "StatsCard", type: "total" }}>
            <StatsCard title="Total Messages" type="total" />
          </ErrorBoundary>
        </div>
        <div>
          <ErrorBoundary context={{ component: "StatsCard", type: "priority" }}>
            <StatsCard title="Priority Breakdown" type="priority" />
          </ErrorBoundary>
        </div>
        <div>
          <ErrorBoundary context={{ component: "StatsCard", type: "type" }}>
            <StatsCard title="Type Breakdown" type="type" />
          </ErrorBoundary>
        </div>
      </div>
    </div>

    <!-- Message Rate Chart - Moved to bottom -->
    <div class="w-full">
      <ErrorBoundary context={{ component: "MessageRateChart" }}>
        <MessageRateChart />
      </ErrorBoundary>
    </div>
  </section>
</main>
