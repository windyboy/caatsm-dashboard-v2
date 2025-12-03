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
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";
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
    loadStats();
    
    // Subscribe to time range changes
    const unsubscribe = timeRange.subscribe(() => {
      loadStats();
    });
    
    return () => {
      unsubscribe();
    };
  });
</script>

<div
  class="min-h-screen text-slate-900 dark:text-slate-100 antialiased bg-gradient-to-br from-blue-50 via-blue-100 to-blue-200 dark:from-slate-900 dark:via-slate-800 dark:to-slate-900"
  style="background-attachment: fixed;"
>
  <header
    class="bg-white/95 dark:bg-slate-800/95 backdrop-blur-lg shadow-lg border-b border-brand-200/30 dark:border-slate-700/50 sticky top-0 z-50"
  >
    <div class="mx-auto max-w-7xl px-4 sm:px-6 py-4 sm:py-5">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2 sm:gap-4">
          <div
            class="w-8 h-8 sm:w-10 sm:h-10 rounded-lg bg-gradient-to-br from-brand-500 via-accent-500 to-success-500 shadow-lg border border-brand-400/30 flex items-center justify-center"
          >
            <svg
              class="w-4 h-4 sm:w-6 sm:h-6 text-white"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
              ></path>
            </svg>
          </div>
          <h1
            class="text-lg sm:text-2xl font-bold text-slate-900 dark:text-slate-100 tracking-tight bg-gradient-to-r from-brand-600 to-accent-600 dark:from-brand-400 dark:to-accent-400 bg-clip-text text-transparent"
          >
            CAATSM Dashboard
          </h1>
        </div>
        <nav class="flex items-center gap-1 sm:gap-2">
          <a
            href="/"
            class="px-3 py-2 sm:px-4 sm:py-2.5 text-xs sm:text-sm font-semibold text-slate-700 dark:text-slate-300 hover:text-brand-600 dark:hover:text-brand-400 hover:bg-brand-50/80 dark:hover:bg-slate-700/80 rounded-lg transition-all duration-300 border border-transparent hover:border-brand-200/60 dark:hover:border-slate-600/60 hover:shadow-md hover:scale-105 flex items-center gap-1 sm:gap-2"
            aria-label="Dashboard"
          >
            <svg
              class="w-3 h-3 sm:w-4 sm:h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
              ></path>
            </svg>
            <span class="hidden xs:inline">Dashboard</span>
          </a>
          <a
            href="/search"
            class="px-3 py-2 sm:px-4 sm:py-2.5 text-xs sm:text-sm font-semibold text-slate-700 dark:text-slate-300 hover:text-brand-600 dark:hover:text-brand-400 hover:bg-brand-50/80 dark:hover:bg-slate-700/80 rounded-lg transition-all duration-300 border border-transparent hover:border-brand-200/60 dark:hover:border-slate-600/60 hover:shadow-md hover:scale-105 flex items-center gap-1 sm:gap-2"
            aria-label="Search"
          >
            <svg
              class="w-3 h-3 sm:w-4 sm:h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              ></path>
            </svg>
            <span class="hidden xs:inline">Search</span>
          </a>
          <ThemeToggle />
        </nav>
      </div>
    </div>
  </header>

  <main class="mx-auto max-w-7xl px-4 sm:px-6 py-6 sm:py-8">
    <section class="space-y-8">
      <!-- Time Range Selector -->
      <div class="w-full">
        <ErrorBoundary context={{ component: "TimeRangeSelector" }}>
          <TimeRangeSelector />
        </ErrorBoundary>
      </div>

      <!-- Message Rate Chart -->
      <div class="w-full">
        <ErrorBoundary context={{ component: "MessageRateChart" }}>
          <MessageRateChart />
        </ErrorBoundary>
      </div>

      <div class="grid gap-6 lg:grid-cols-3">
        <!-- Live Stream - Full width on mobile, 2/3 on desktop -->
        <div class="lg:col-span-2 order-2 lg:order-1">
          <ErrorBoundary context={{ component: "LiveStream" }}>
            <LiveStream />
          </ErrorBoundary>
        </div>

        <!-- Stats Cards - Stack on mobile, sidebar on desktop -->
        <div class="space-y-4 order-1 lg:order-2">
          <div class="grid gap-4 sm:grid-cols-1 md:grid-cols-3 lg:grid-cols-1">
            <div class="md:col-span-1 lg:col-span-1">
              <ErrorBoundary context={{ component: "HealthStatus" }}>
                <HealthStatus />
              </ErrorBoundary>
            </div>
            <div class="md:col-span-1 lg:col-span-1">
              <ErrorBoundary context={{ component: "MetricsCard" }}>
                <MetricsCard />
              </ErrorBoundary>
            </div>
            <div class="md:col-span-1 lg:col-span-1">
              <ErrorBoundary context={{ component: "WebSocketMetrics" }}>
                <WebSocketMetrics />
              </ErrorBoundary>
            </div>
            <div class="md:col-span-1 lg:col-span-1">
              <ErrorBoundary context={{ component: "StatsCard", type: "total" }}>
                <StatsCard title="Total Messages" type="total" />
              </ErrorBoundary>
            </div>
            <div class="md:col-span-1 lg:col-span-1">
              <ErrorBoundary context={{ component: "StatsCard", type: "priority" }}>
                <StatsCard title="Priority Breakdown" type="priority" />
              </ErrorBoundary>
            </div>
            <div class="md:col-span-1 lg:col-span-1">
              <ErrorBoundary context={{ component: "StatsCard", type: "type" }}>
                <StatsCard title="Type Breakdown" type="type" />
              </ErrorBoundary>
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</div>
