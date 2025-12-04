<script lang="ts">
  import { onMount } from "svelte";
  import HealthCard from "$lib/components/HealthCard.svelte";
  import MessageList from "$lib/components/MessageList.svelte";
  import MetricCard from "$lib/components/MetricCard.svelte";
  import LoadingSkeleton from "$lib/components/LoadingSkeleton.svelte";
  import { fetchHealth, fetchRecentMessages, fetchStats } from "$lib/api";
  import { getWebSocketStore } from "$lib/stores/websocket.svelte";
  import type { HealthSnapshot, TrafficSummary, Telegram, SearchResponse } from "$lib/types";

  const wsStore = getWebSocketStore();
  // Initial HTTP data (only used as fallback until WebSocket connects)
  let initialStats = $state<TrafficSummary | null>(null);
  let initialHealth = $state<HealthSnapshot | null>(null);
  let initialMessages = $state<Telegram[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

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

  const displayHealth = $derived(wsStore.health ?? initialHealth);

  // Messages: WebSocket is primary, HTTP only as initial fallback
  const displayMessages = $derived.by(() => {
    const wsMessages = wsStore.newMessages;
    // Use WebSocket messages if available (they are always the most up-to-date)
    if (wsMessages.length > 0) {
      return wsMessages.slice(0, 50);
    }
    // Fallback to initial HTTP messages only if WebSocket is not connected or empty
    return initialMessages.slice(0, 50);
  });

  const healthItems = $derived.by(() => {
    const health = displayHealth;
    if (!health) return [];
    
    const getDetail = (component: { status?: string; message?: string; response_time?: string } | undefined): string | undefined => {
      if (!component) return undefined;
      // For error status, prefer message; otherwise prefer response_time
      if (component.status === "error" && component.message) {
        return component.message;
      }
      return component.response_time || component.message;
    };
    
    // Build items array - always include all components that exist in the health object
    const items: { label: string; status?: string; detail?: string }[] = [];
    
    // Check each component and add it if it exists in the health object
    if (health.postgresql != null) {
      items.push({ 
        label: "PostgreSQL", 
        status: health.postgresql.status || "unknown", 
        detail: getDetail(health.postgresql) 
      });
    }
    if (health.redis != null) {
      items.push({ 
        label: "Redis", 
        status: health.redis.status || "unknown", 
        detail: getDetail(health.redis) 
      });
    }
    if (health.meilisearch != null) {
      items.push({ 
        label: "Meilisearch", 
        status: health.meilisearch.status || "unknown", 
        detail: getDetail(health.meilisearch) 
      });
    }
    if (health.nats != null) {
      items.push({ 
        label: "NATS", 
        status: health.nats.status || "unknown", 
        detail: getDetail(health.nats) 
      });
    }
    
    return items;
  });

  const activeRoutes = $derived(displayStats?.activeRoutes ?? null);
  const messagesPerSec = $derived(displayStats?.messagesPerSec ?? null);

  const typeDistribution = $derived.by(() => {
    const byType = displayStats?.byType;
    if (!byType || Object.keys(byType).length === 0) return "—";
    const total = displayStats?.total ?? 0;
    return Object.entries(byType)
      .map(([type, count]) => {
        const percentage = total > 0 ? Math.round((count / total) * 100) : 0;
        return `${type}: ${count.toLocaleString()} (${percentage}%)`;
      })
      .join(" | ");
  });

  const timeWindowLabel = $derived.by(() => {
    const window = displayStats?.timeWindow;
    if (!window) return "All time";
    switch (window) {
      case "last_1h":
        return "Last 1 hour";
      case "last_24h":
        return "Last 24 hours";
      default:
        return "All time";
    }
  });

  onMount(async () => {
    document.title = "CAATSM Dashboard - Operations";
    wsStore.connect();

    loading = true;
    error = null;

    try {
      const [statsResponse, healthResponse, messagesResponse] = await Promise.all([
        fetchStats().catch(() => null),
        fetchHealth().catch((err) => {
          console.error("Failed to fetch health from /api/health:", err);
          return null;
        }),
        fetchRecentMessages(12).catch(() => ({ telegrams: [], total: 0 } as SearchResponse)),
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

      if (!statsResponse && !healthResponse && (!messagesResponse || messagesResponse.telegrams.length === 0)) {
        error = "Unable to connect to backend API. Please ensure the backend server is running on port 3002. You can start it with: make backend-dev";
      }
    } catch (err) {
      error = err instanceof Error ? err.message : "Unable to load dashboard data.";
    } finally {
      loading = false;
    }
  });

</script>

<main class="page">
  <section class="page__intro">
    <p class="eyebrow">Live operational view</p>
    <h1>Messaging health at a glance</h1>
    <p class="muted">
      Simplified dashboard that highlights the totals, health signals, and latest messages coming from
      the backend.
    </p>
  </section>

  <section class="grid grid--metrics">
    <MetricCard title="Messages" value={displayStats?.total?.toLocaleString() ?? "—"} hint={timeWindowLabel} />
    <MetricCard
      title="Messages/sec"
      value={messagesPerSec !== null ? messagesPerSec.toFixed(1) : "—"}
      hint="Last 1 min"
    />
    <MetricCard
      title="Active routes"
      value={activeRoutes !== null ? activeRoutes.toLocaleString() : "—"}
      hint={`Unique source→destination pairs (${timeWindowLabel.toLowerCase()})`}
    />
  </section>

  {#if typeDistribution && typeDistribution !== "—"}
    <div class="card">
      <p class="eyebrow">Type Distribution ({timeWindowLabel})</p>
      <p class="muted">{typeDistribution}</p>
    </div>
  {/if}

  {#if wsStore.status !== "disconnected"}
    <div class="card ws-status">
      <span class="status-dot {wsStore.status === 'connected' ? 'pill--ok' : wsStore.status === 'connecting' ? 'pill--warn' : 'pill--error'}"></span>
      <span class="muted">
        {wsStore.status === "connected" ? "Real-time updates active" : wsStore.status === "connecting" ? "Connecting..." : "Connection error"}
      </span>
    </div>
  {/if}

  {#if loading}
    <section class="grid grid--metrics">
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
    <section class="grid grid--two">
      <HealthCard status={displayHealth?.status} items={healthItems} />
      <MessageList messages={displayMessages} />
    </section>
  {/if}
</main>
