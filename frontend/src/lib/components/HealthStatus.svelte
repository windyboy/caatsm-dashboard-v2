<!--
  @component HealthStatus
  Displays system health status for all components (PostgreSQL, Redis, Meilisearch, NATS).
  
  Features:
  - Real-time health status updates (polling every 30 seconds)
  - Color-coded status indicators (green=ok, yellow=warning, red=error)
  - Last check timestamp
-->

<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { getHealthStatus, type HealthCheckResult, type ComponentHealth } from "$lib/services/api";
  import { createLogger } from "$lib/utils/logger";

  const logger = createLogger("HealthStatus");

  let healthData = $state<HealthCheckResult | null>(null);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let lastCheckTime = $state<Date | null>(null);
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  const POLL_INTERVAL_MS = 30000; // 30 seconds

  function getStatusColor(status: string): string {
    switch (status) {
      case "ok":
        return "text-green-600 bg-green-50 border-green-200";
      case "error":
        return "text-red-600 bg-red-50 border-red-200";
      case "not_configured":
        return "text-yellow-600 bg-yellow-50 border-yellow-200";
      default:
        return "text-gray-600 bg-gray-50 border-gray-200";
    }
  }

  function getStatusIcon(status: string): string {
    switch (status) {
      case "ok":
        return "✓";
      case "error":
        return "✗";
      case "not_configured":
        return "○";
      default:
        return "?";
    }
  }

  function getStatusLabel(status: string): string {
    switch (status) {
      case "ok":
        return "Healthy";
      case "error":
        return "Error";
      case "not_configured":
        return "Not Configured";
      default:
        return "Unknown";
    }
  }

  async function fetchHealthStatus() {
    if (loading) return;

    loading = true;
    error = null;

    try {
      const result = await getHealthStatus();
      healthData = result;
      lastCheckTime = new Date();
      logger.debug("Health status fetched", { status: result.status });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to fetch health status";
      error = errorMessage;
      logger.error("Failed to fetch health status", err);
    } finally {
      loading = false;
    }
  }

  function startPolling() {
    // Initial fetch
    fetchHealthStatus();

    // Set up polling
    pollInterval = setInterval(() => {
      fetchHealthStatus();
    }, POLL_INTERVAL_MS);
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
  }

  onMount(() => {
    if (typeof window !== "undefined") {
      startPolling();
    }
  });

  onDestroy(() => {
    stopPolling();
  });
</script>

<div class="health-status-card" role="region" aria-label="System Health Status">
  <div class="health-header">
    <div class="health-icon">
      <svg class="w-4 h-4 text-brand-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
        ></path>
      </svg>
    </div>
    <p class="health-title">System Health</p>
  </div>

  {#if loading && !healthData}
    <div class="health-loading">
      <p class="text-sm text-slate-500">Checking health status...</p>
    </div>
  {:else if error && !healthData}
    <div class="health-error">
      <p class="text-sm font-semibold text-red-600 mb-1">Error</p>
      <p class="text-xs text-slate-500">{error}</p>
    </div>
  {:else if healthData}
    <div class="health-content">
      <!-- Overall Status -->
      <div class="overall-status">
        <span class="status-badge {getStatusColor(healthData.status)}">
          {getStatusIcon(healthData.status)} {healthData.status === "ok" ? "All Systems Operational" : "System Degraded"}
        </span>
      </div>

      <!-- Component Status List -->
      {#if healthData.checks && typeof healthData.checks === 'object'}
        <div class="component-list">
          {#each Object.entries(healthData.checks) as [name, component]}
            {@const componentHealth = component as ComponentHealth}
            <div class="component-item">
              <div class="component-name">
                <span class="component-icon {getStatusColor(componentHealth.status)}">
                  {getStatusIcon(componentHealth.status)}
                </span>
                <span class="component-label">{name.charAt(0).toUpperCase() + name.slice(1)}</span>
              </div>
              <div class="component-status">
                <span class="status-text {getStatusColor(componentHealth.status)}">
                  {getStatusLabel(componentHealth.status)}
                </span>
                {#if componentHealth.message}
                  <span class="component-message" title={componentHealth.message}>
                    ⓘ
                  </span>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {:else}
        <div class="component-list">
          <p class="text-xs text-slate-500">No component health data available</p>
        </div>
      {/if}

      <!-- Last Check Time -->
      {#if lastCheckTime}
        <div class="last-check">
          <p class="text-xs text-slate-400">
            Last checked: {lastCheckTime.toLocaleTimeString()}
          </p>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .health-status-card {
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

  .health-status-card:hover {
    background: var(--stats-card-hover-bg);
    box-shadow: var(--stats-card-hover-shadow);
    transform: scale(1.05);
  }

  .health-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 0.75rem;
  }

  .health-icon {
    padding: 0.5rem;
    border-radius: 0.5rem;
    border: 1px solid rgba(191, 219, 254, 0.5);
    background: linear-gradient(to right, #dbeafe, #f3e8ff);
  }

  .health-title {
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

  .health-loading,
  .health-error {
    text-align: center;
    padding: 1rem 0;
  }

  .health-content {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .overall-status {
    display: flex;
    justify-content: center;
    margin-bottom: 0.5rem;
  }

  .status-badge {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 600;
    border: 1px solid;
  }

  .component-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .component-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5rem;
    border-radius: 0.375rem;
    background: rgba(248, 250, 252, 0.5);
  }

  .component-name {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .component-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    font-weight: 600;
    border: 1px solid;
  }

  .component-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #334155;
    text-transform: capitalize;
  }

  .component-status {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .status-text {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid;
  }

  .component-message {
    font-size: 0.75rem;
    color: #64748b;
    cursor: help;
  }

  .last-check {
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
    text-align: center;
  }
</style>

