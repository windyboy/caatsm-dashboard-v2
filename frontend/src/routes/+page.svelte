<script lang="ts">
  import { onMount } from "svelte";
  import HealthCard from "$lib/components/HealthCard.svelte";
  import MessageList from "$lib/components/MessageList.svelte";
  import MetricCard from "$lib/components/MetricCard.svelte";
  import { fetchHealth, fetchRecentMessages, fetchStats } from "$lib/api";
  import type { HealthSnapshot, TrafficSummary, Telegram, SearchResponse } from "$lib/types";

  let stats = $state<TrafficSummary | null>(null);
  let health = $state<HealthSnapshot | null>(null);
  let messages = $state<Telegram[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  const healthItems = $derived(
    [
      { label: "PostgreSQL", status: health?.postgresql?.status, detail: health?.postgresql?.message },
      { label: "Redis", status: health?.redis?.status, detail: health?.redis?.message },
      { label: "Meilisearch", status: health?.meilisearch?.status, detail: health?.meilisearch?.message },
      { label: "NATS", status: health?.nats?.status, detail: health?.nats?.message },
    ].filter((item) => item.status || item.detail)
  );

  const totalTypes = $derived(
    stats?.byType ? Object.keys(stats.byType).length : 0
  );
  const totalPriorities = $derived(
    stats?.byPriority ? Object.keys(stats.byPriority).length : 0
  );

  onMount(async () => {
    loading = true;
    error = null;

    try {
      const [statsResponse, healthResponse, messagesResponse] = await Promise.all([
        fetchStats().catch((err) => {
          console.error("Failed to fetch stats from /api/stats:", err);
          return null;
        }),
        fetchHealth().catch((err) => {
          console.error("Failed to fetch health from /api/health:", err);
          return null;
        }),
        fetchRecentMessages(12).catch((err) => {
          console.error("Failed to fetch messages from /api/search:", err);
          return { telegrams: [], total: 0 } as SearchResponse;
        }),
      ]);

      if (statsResponse) {
        stats = statsResponse;
      }
      if (healthResponse) {
        health = healthResponse;
      }
      if (messagesResponse) {
        messages = messagesResponse.telegrams ?? [];
      }

      // If all requests failed, show error with helpful message
      if (!statsResponse && !healthResponse && (!messagesResponse || messagesResponse.telegrams.length === 0)) {
        error = "Unable to connect to backend API. Please ensure the backend server is running on port 3002. You can start it with: make backend-dev";
      }
    } catch (err) {
      console.error("Dashboard load error:", err);
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
    <MetricCard title="Messages (last window)" value={stats?.total ?? "—"} hint="From /api/stats" />
    <MetricCard title="Priority buckets" value={totalPriorities > 0 ? totalPriorities : "—"} hint="Count of priority groups" />
    <MetricCard title="Types tracked" value={totalTypes > 0 ? totalTypes : "—"} hint="By message type" />
  </section>

  {#if loading}
    <div class="card">
      <p class="eyebrow">Loading</p>
      <p class="muted">Fetching the latest numbers...</p>
    </div>
  {:else if error}
    <div class="card error-card">
      <p class="eyebrow">Error</p>
      <p class="muted">{error}</p>
    </div>
  {:else}
    <section class="grid grid--two">
      <HealthCard status={health?.status} items={healthItems} />
      <MessageList {messages} />
    </section>
  {/if}
</main>
