<!--
  @component HealthStatus
  Displays system health status for all components (PostgreSQL, Redis, Meilisearch, NATS).
  
  Features:
  - Real-time health status updates (polling every 30 seconds)
  - Color-coded status indicators (green=ok, yellow=warning, red=error)
  - Last check timestamp
  - Version and uptime information
-->

<script lang="ts">
  import type { ComponentHealth } from "$lib/services/api";
  import { useHealthPolling } from "$lib/composables/useHealthPolling.svelte";
  import { formatUptime } from "$lib/utils/health";
  import HealthStatusIndicator from "./HealthStatusIndicator.svelte";
  import ComponentHealthList from "./ComponentHealthList.svelte";

  const { healthData, loading, error, lastCheckTime } = useHealthPolling();
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
        <HealthStatusIndicator status={healthData.status} />
      </div>

      <!-- Version and Uptime Info -->
      {#if healthData.version || healthData.uptime}
        <div class="system-info">
          {#if healthData.version}
            <span class="info-item">
              <span class="info-label">Version:</span>
              <span class="info-value">{healthData.version}</span>
            </span>
          {/if}
          {#if healthData.uptime}
            <span class="info-item">
              <span class="info-label">Uptime:</span>
              <span class="info-value">{formatUptime(healthData.uptime)}</span>
            </span>
          {/if}
        </div>
      {/if}

      <!-- Component Status List -->
      <ComponentHealthList
        components={[
          { name: "PostgreSQL", key: "postgresql", health: healthData.postgresql },
          { name: "Meilisearch", key: "meilisearch", health: healthData.meilisearch },
          { name: "Redis", key: "redis", health: healthData.redis },
          { name: "NATS", key: "nats", health: healthData.nats },
        ]}
      />

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

  .system-info {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    justify-content: center;
    padding: 0.5rem;
    border-radius: 0.375rem;
    background: rgba(248, 250, 252, 0.5);
  }

  .info-item {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
  }

  .info-label {
    color: #64748b;
    font-weight: 500;
  }

  .info-value {
    color: #334155;
    font-weight: 600;
  }

  .last-check {
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid rgba(148, 163, 184, 0.2);
    text-align: center;
  }
</style>
