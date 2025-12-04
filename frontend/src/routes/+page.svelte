<script lang="ts">
  import { onMount } from "svelte";
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
  import type { SortField, SortOrder } from "$lib/components/MessageDataTable.svelte";

  const wsStore = getWebSocketStore();
  // Initial HTTP data (only used as fallback until WebSocket connects)
  let initialStats = $state<TrafficSummary | null>(null);
  let initialHealth = $state<HealthSnapshot | null>(null);
  let initialMessages = $state<Telegram[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let timeWindow = $state("last_24h");

  // WebSocket data is primary source - simple priority: WS > HTTP fallback
  const displayStats = $derived.by(() => {
    const wsStats = wsStore.stats;
    if (wsStats) {
      return {
        total: wsStats.total,
        byType: wsStats.byType,
        activeRoutes: wsStats.activeRoutes,
        messagesPerSec: wsStats.messagesPerSec,
        timeWindow: wsStats.timeWindow,
      } as TrafficSummary;
    }
    // Fallback to initial HTTP data
    return initialStats;
  });

  const displayHealth = $derived<HealthSnapshot | null>(wsStore.health ?? initialHealth);

  // Messages: WebSocket is primary, HTTP only as initial fallback
  const displayMessages = $derived.by((): Telegram[] => {
    const wsMessages = wsStore.newMessages;
    // Use WebSocket messages if available (they are always the most up-to-date)
    if (wsMessages.length > 0) {
      return wsMessages.slice(0, 100);
    }
    // Fallback to initial HTTP messages only if WebSocket is not connected or empty
    return initialMessages.slice(0, 100);
  });

  const activeRoutes = $derived(displayStats?.activeRoutes ?? null);
  const messagesPerSec = $derived(displayStats?.messagesPerSec ?? null);

  const timeWindowLabel = $derived.by(() => {
    const window = displayStats?.timeWindow || timeWindow;
    switch (window) {
      case "last_1h":
        return "Last 1 hour";
      case "last_24h":
        return "Last 24 hours";
      case "last_7d":
        return "Last 7 days";
      default:
        return "All time";
    }
  });

  // Mock trend data (can be replaced with real data later)
  const messagesTrend = $derived.by(() => {
    // In real implementation, compare with previous period
    return { value: 12.5, label: "Trending up this month" };
  });

  function handleTimeWindowChange(value: string) {
    timeWindow = value;
    // TODO: Trigger data refetch with new time window
  }

  function handleRefresh() {
    // Reload data
    window.location.reload();
  }

  function handleSort(field: SortField, order: SortOrder) {
    // Sorting is handled by MessageDataTable component
  }

  function handlePageChange(page: number) {
    // Pagination is handled by MessageDataTable component
  }

  onMount(async () => {
    document.title = "CAATSM Dashboard - Operations";
    wsStore.connect();

    loading = true;
    error = null;

    try {
      const [statsResponse, healthResponse, messagesResponse] = await Promise.all([
        fetchStats().catch(() => null),
        fetchHealth().catch(() => null),
        fetchRecentMessages(50).catch(() => ({ telegrams: [], total: 0 }) as SearchResponse),
      ]);

      // Store initial HTTP data as fallback (only used until WebSocket connects and provides data)
      if (statsResponse) {
        initialStats = statsResponse;
      }
      if (healthResponse) {
        initialHealth = healthResponse;
      }
      if (messagesResponse) {
        initialMessages = messagesResponse.telegrams ?? [];
      }

      // Connection errors are handled by ConnectionBanner
      // Only set error for non-connection related issues
      // Even if all requests fail, we still show the UI with empty data
      // so components can render properly
    } catch (err) {
      // Connection errors are handled by ConnectionBanner
      // Only set error for non-connection related issues
      if (
        err instanceof Error &&
        !err.message.includes("无法连接") &&
        !err.message.includes("Network error")
      ) {
        error = err.message;
      }
    } finally {
      // Always set loading to false so components can render
      // Components should handle empty/null data gracefully
      loading = false;
    }
  });
</script>

<main class="dashboard-page">
  <ConnectionBanner />

  <TopControlBar
    {timeWindow}
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
    {:else if error}
      <div class="card error-card" role="alert" aria-live="polite" aria-atomic="true">
        <p class="eyebrow">Error</p>
        <p class="muted">{error}</p>
      </div>
    {:else}
      <!-- KPI Metrics Grid -->
      <section class="grid grid--four">
        <EnhancedMetricCard
          title="Total Messages"
          value={displayStats?.total?.toLocaleString() ?? "—"}
          hint={timeWindowLabel}
          trend={messagesTrend}
          variant="primary"
        />
        <EnhancedMetricCard
          title="Messages/sec"
          value={messagesPerSec !== null ? messagesPerSec.toFixed(1) : "—"}
          hint="Last 1 min"
        />
        <EnhancedMetricCard
          title="Active Routes"
          value={activeRoutes !== null ? activeRoutes.toLocaleString() : "—"}
          hint={`Unique source→destination pairs`}
        />
        <EnhancedMetricCard
          title="System Health"
          value={displayHealth?.status?.toUpperCase() ?? "—"}
          hint="Overall status"
          variant="default"
        />
      </section>

      <!-- Visualization Section -->
      <section class="grid grid--two">
        <TypeDistributionChart data={displayStats?.byType} title="Message Type Distribution" />
        <MessageTrendChart title="Message Trend" />
      </section>

      <!-- System Health Panel -->
      <section class="health-section">
        <HealthStatusGrid health={displayHealth || null} />
      </section>

      <!-- Real-time Message Table -->
      <section class="messages-section">
        <MessageDataTable
          messages={displayMessages}
          pageSize={10}
          {loading}
          onSort={handleSort}
          onPageChange={handlePageChange}
        />
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

  .error-card {
    border-color: #fecdd3;
    background: #fff1f2;
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
