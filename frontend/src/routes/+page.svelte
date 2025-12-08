<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import TopControlBar from "$lib/components/TopControlBar.svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import ConnectionBanner from "$lib/components/ConnectionBanner.svelte";
  import KPIMetricsSection from "$lib/components/dashboard/KPIMetricsSection.svelte";
  import VisualizationSection from "$lib/components/dashboard/VisualizationSection.svelte";
  import HealthSection from "$lib/components/dashboard/HealthSection.svelte";
  import MessagesSection from "$lib/components/dashboard/MessagesSection.svelte";
  import { fetchHealth, fetchRecentMessages, fetchStats, fetchTrendData } from "$lib/api";
  import { getWebSocketStore } from "$lib/stores/websocket.svelte";
  import type {
    HealthSnapshot,
    TrafficSummary,
    Telegram,
    SearchResponse,
  } from "$lib/types";
  import type { TrendDataPoint } from "$lib/components/charts/MessageTrendChart.svelte";

  // Simple state management - no complex derived values
  let stats = $state<TrafficSummary | null>(null);
  let health = $state<HealthSnapshot | null>(null);
  let messages = $state<Telegram[]>([]);
  let trendData = $state<TrendDataPoint[]>([]);
  let loading = $state(true);
  let timeWindow = $state("last_24h");

  // Initialize WebSocket store
  const wsStore = getWebSocketStore();

  // Simple derived values - no nested chains
  const totalMessages = $derived(stats?.total ?? 0);
  const messagesPerSec = $derived(stats?.messagesPerSec ?? 0);
  const activeRoutes = $derived(stats?.activeRoutes ?? 0);
  const byType = $derived(stats?.byType ?? {});
  const systemStatus = $derived(health?.status ?? "unknown");

  async function handleTimeWindowChange(value: string) {
    timeWindow = value;
    // Reload data with new time window
    loading = true;
    try {
      const [statsResponse, trendResponse] = await Promise.all([
        fetchStats(value).catch(() => null),
        fetchTrendData(value).catch(() => null),
      ]);

      if (statsResponse) {
        console.log("Stats response for time window:", value, statsResponse);
        stats = {
          ...statsResponse,
          byType: statsResponse.byType ?? {},
        };
        console.log("Updated stats with byType:", stats.byType);
      }
      if (trendResponse?.data) {
        trendData = trendResponse.data.map((d) => ({
          time: d.time,
          count: d.count,
        }));
      }
    } catch (err) {
      console.error("Failed to load data for time window:", err);
    } finally {
      loading = false;
    }
  }

  function handleRefresh() {
    // Reset loading state and refetch data
    loading = true;
    Promise.all([
      fetchStats(timeWindow).catch(() => null),
      fetchHealth().catch(() => null),
      fetchRecentMessages(50).catch(() => ({ telegrams: [], total: 0 }) as SearchResponse),
      fetchTrendData(timeWindow).catch(() => null),
    ]).then(([statsResponse, healthResponse, messagesResponse, trendResponse]) => {
      if (statsResponse) {
        stats = {
          ...statsResponse,
          byType: statsResponse.byType ?? {},
        };
      }
      if (healthResponse) health = healthResponse;
      if (messagesResponse?.telegrams) messages = messagesResponse.telegrams;
      if (trendResponse?.data) {
        trendData = trendResponse.data.map((d) => ({
          time: d.time,
          count: d.count,
        }));
      }
    }).catch((err) => {
      console.error("Failed to refresh data:", err);
    }).finally(() => {
      loading = false;
    });
  }

  // Reactive sync of WebSocket data to local state using $effect
  // This replaces polling with reactive updates for better performance
  // Note: Only sync stats if WebSocket timeWindow matches current selection
  // to avoid overriding manual time window changes
  $effect(() => {
    const wsStats = wsStore.stats;
    const wsHealth = wsStore.health;
    const wsMessages = wsStore.newMessages;

    // Sync stats - only update from WebSocket if we're viewing "last_24h"
    // WebSocket always sends "last_24h" data, so we should only sync when that's selected
    // or if we don't have any stats yet (initial connection)
    if (wsStats && (timeWindow === "last_24h" || !stats)) {
        const newByType = wsStats.byType ?? {};
        
        // Check if we need to update by comparing values first
        const needsUpdate = !stats ||
          stats.total !== wsStats.total ||
          stats.activeRoutes !== (wsStats.activeRoutes ?? 0) ||
          stats.messagesPerSec !== (wsStats.messagesPerSec ?? 0) ||
          stats.timeWindow !== wsStats.timeWindow;
        
        // Check if byType changed (deep comparison to avoid reference issues)
        let byTypeChanged = false;
        if (stats) {
          const oldByType = stats.byType ?? {};
          const oldKeys = Object.keys(oldByType);
          const newKeys = Object.keys(newByType);
          byTypeChanged = oldKeys.length !== newKeys.length ||
            oldKeys.some(key => (oldByType[key] ?? 0) !== (newByType[key] ?? 0));
        } else {
          // If stats is null, check if newByType has any data
          byTypeChanged = Object.keys(newByType).length > 0;
        }
        
      if (needsUpdate || byTypeChanged) {
        stats = {
          total: wsStats.total,
          byType: newByType,
          activeRoutes: wsStats.activeRoutes ?? 0,
          messagesPerSec: wsStats.messagesPerSec ?? 0,
          timeWindow: wsStats.timeWindow,
        };
      }
    }

    // Sync health - only update if changed
    if (wsHealth && wsHealth !== health) {
      health = wsHealth;
    }

    // Sync messages - use slice to create new array reference only when needed
    if (wsMessages.length !== messages.length || 
        wsMessages[0]?.message_id !== messages[0]?.message_id) {
      messages = wsMessages.slice(0, 100);
    }
  });

  onMount(async () => {
    document.title = "CAATSM Dashboard - Operations";

    // Connect WebSocket
    wsStore.connect();

    // Initial HTTP data fetch
    try {
      const [statsResponse, healthResponse, messagesResponse, trendResponse] = await Promise.all([
        fetchStats(timeWindow).catch(() => null),
        fetchHealth().catch(() => null),
        fetchRecentMessages(50).catch(() => ({ telegrams: [], total: 0 }) as SearchResponse),
        fetchTrendData(timeWindow).catch(() => null),
      ]);

      if (statsResponse) {
        stats = {
          ...statsResponse,
          byType: statsResponse.byType ?? {},
        };
      }
      if (healthResponse) health = healthResponse;
      if (messagesResponse?.telegrams) messages = messagesResponse.telegrams;
      if (trendResponse?.data) {
        trendData = trendResponse.data.map((d) => ({
          time: d.time,
          count: d.count,
        }));
      }
    } catch (err) {
      console.error("Failed to load initial data:", err);
    } finally {
      loading = false;
    }
  });

  onDestroy(() => {
    // Clean up WebSocket connection when component is destroyed
    wsStore.disconnect();
  });
</script>

<div class="dashboard-page">
  <ConnectionBanner />

  <TopControlBar
    bind:timeWindow
    onTimeWindowChange={handleTimeWindowChange}
    onRefresh={handleRefresh}
  />

  <div class="dashboard-content">
    <section class="page__intro">
      <p class="eyebrow">Live operational view</p>
      <h1>Messaging health at a glance</h1>
      <p class="muted">
        Real-time dashboard tracking aviation telegram traffic across AFTN, SITA, ACARS, and CPDLC
        networks.
      </p>
    </section>

    {#if loading}
      <section class="grid grid--four">
        <LoadingSkeleton variant="metric" />
        <LoadingSkeleton variant="metric" />
        <LoadingSkeleton variant="metric" />
        <LoadingSkeleton variant="metric" />
      </section>
      <section class="grid grid--two">
        <LoadingSkeleton variant="health" />
        <LoadingSkeleton variant="message" />
      </section>
    {:else}
      <KPIMetricsSection
        {totalMessages}
        {messagesPerSec}
        {activeRoutes}
        {systemStatus}
      />

      <VisualizationSection {byType} {trendData} />

      <HealthSection {health} />

      <MessagesSection {messages} {loading} />
    {/if}
  </div>
</div>

<style>
  .dashboard-page {
    min-height: 100vh;
    background: #fafafa;
  }

  .dashboard-content {
    max-width: 1400px;
    margin: 0 auto;
    padding: 1.5rem;
  }

  .page__intro {
    margin-bottom: 2rem;
  }

  .page__intro h1 {
    margin: 0.35rem 0 0.5rem;
    font-size: 2rem;
    font-weight: 700;
    color: #18181b;
  }

  .grid {
    display: grid;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .grid--four {
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  }

  .grid--two {
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  }

  @media (max-width: 1024px) {
    .grid--four {
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    }

    .grid--two {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .dashboard-content {
      padding: 1rem;
    }

    .grid--four {
      grid-template-columns: 1fr;
    }

    .page__intro h1 {
      font-size: 1.5rem;
    }
  }
</style>
