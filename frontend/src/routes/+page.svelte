<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import TopControlBar from "$lib/components/TopControlBar.svelte";
  import EnhancedMetricCard from "$lib/components/EnhancedMetricCard.svelte";
  import TypeDistributionChart from "$lib/components/charts/TypeDistributionChart.svelte";
  import MessageTrendChart from "$lib/components/charts/MessageTrendChart.svelte";
  import HealthStatusGrid from "$lib/components/HealthStatusGrid.svelte";
  import MessageDataTable from "$lib/components/MessageDataTable.svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import ConnectionBanner from "$lib/components/ConnectionBanner.svelte";
  import { fetchHealth, fetchRecentMessages, fetchStats } from "$lib/api";
  import { getWebSocketStore } from "$lib/stores/websocket.svelte";
  import type { HealthSnapshot, TrafficSummary, Telegram, SearchResponse } from "$lib/types";

  // Simple state management - no complex derived values
  let stats = $state<TrafficSummary | null>(null);
  let health = $state<HealthSnapshot | null>(null);
  let messages = $state<Telegram[]>([]);
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

  function handleTimeWindowChange(value: string) {
    timeWindow = value;
  }

  function handleRefresh() {
    window.location.reload();
  }

  // Reactive sync of WebSocket data to local state using $effect
  // This replaces polling with reactive updates for better performance
  $effect(() => {
    // Sync stats
    if (wsStore.stats) {
      stats = {
        total: wsStore.stats.total,
        byType: wsStore.stats.byType,
        activeRoutes: wsStore.stats.activeRoutes ?? 0,
        messagesPerSec: wsStore.stats.messagesPerSec ?? 0,
        timeWindow: wsStore.stats.timeWindow,
      };
    }

    // Sync health
    if (wsStore.health) {
      health = wsStore.health;
    }

    // Sync messages - always sync, even if empty, to ensure consistency
    messages = wsStore.newMessages.slice(0, 100);
  });

  onMount(async () => {
    document.title = "CAATSM Dashboard - Operations";
    
    // Connect WebSocket
    wsStore.connect();

    // Initial HTTP data fetch
    try {
      const [statsResponse, healthResponse, messagesResponse] = await Promise.all([
        fetchStats().catch(() => null),
        fetchHealth().catch(() => null),
        fetchRecentMessages(50).catch(() => ({ telegrams: [], total: 0 }) as SearchResponse),
      ]);

      if (statsResponse) stats = statsResponse;
      if (healthResponse) health = healthResponse;
      if (messagesResponse?.telegrams) messages = messagesResponse.telegrams;
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

<main class="dashboard-page">
  <ConnectionBanner />

  <TopControlBar {timeWindow} onTimeWindowChange={handleTimeWindowChange} onRefresh={handleRefresh} />

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
      <!-- KPI Metrics Grid -->
      <section class="grid grid--four">
        <EnhancedMetricCard
          title="Total Messages"
          value={totalMessages.toLocaleString()}
          hint="Last 24 hours"
        />
        <EnhancedMetricCard
          title="Messages/sec"
          value={messagesPerSec.toFixed(1)}
          hint="Last 1 min"
        />
        <EnhancedMetricCard
          title="Active Routes"
          value={activeRoutes.toLocaleString()}
          hint="Unique source→destination pairs"
        />
        <EnhancedMetricCard
          title="System Health"
          value={systemStatus.toUpperCase()}
          hint="Overall status"
        />
      </section>

      <!-- Visualization Section -->
      <section class="grid grid--two">
        <TypeDistributionChart data={byType} title="Message Type Distribution" />
        <MessageTrendChart title="Message Trend" />
      </section>

      <!-- System Health Panel -->
      <section class="health-section">
        <HealthStatusGrid {health} />
      </section>

      <!-- Real-time Message Table -->
      <section class="messages-section">
        <MessageDataTable {messages} pageSize={10} {loading} />
      </section>
    {/if}
  </div>
</main>

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

  .health-section {
    margin-bottom: 1.5rem;
  }

  .messages-section {
    margin-bottom: 2rem;
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
