<script lang="ts">
  import Card from "./ui/Card.svelte";
  import Badge from "./ui/Badge.svelte";
  import { cn } from "./ui/utils";
  import type { HealthSnapshot } from "$lib/types";

  interface HealthStatusGridProps {
    health?: HealthSnapshot | null;
  }

  let { health }: HealthStatusGridProps = $props();

  const statusClass = (value?: string): "ok" | "warn" | "error" | "idle" => {
    if (!value) return "idle";
    const normalized = value.toLowerCase();
    if (["ok", "healthy", "connected"].includes(normalized)) return "ok";
    if (["degraded", "warning"].includes(normalized)) return "warn";
    if (["unhealthy", "error"].includes(normalized)) return "error";
    if (["not_configured"].includes(normalized)) return "idle";
    return "idle";
  };

  const components = $derived.by(() => {
    const items: Array<{
      label: string;
      status: string;
      detail?: string;
      responseTime?: string;
    }> = [];

    if (health?.postgresql) {
      items.push({
        label: "PostgreSQL",
        status: health.postgresql.status || "unknown",
        detail: health.postgresql.message,
        responseTime: health.postgresql.response_time,
      });
    }

    if (health?.redis) {
      items.push({
        label: "Redis",
        status: health.redis.status || "unknown",
        detail: health.redis.message,
        responseTime: health.redis.response_time,
      });
    }

    if (health?.meilisearch) {
      items.push({
        label: "Meilisearch",
        status: health.meilisearch.status || "unknown",
        detail: health.meilisearch.message,
        responseTime: health.meilisearch.response_time,
      });
    }

    if (health?.nats) {
      items.push({
        label: "NATS",
        status: health.nats.status || "unknown",
        detail: health.nats.message,
        responseTime: health.nats.response_time,
      });
    }

    return items;
  });

  const overallStatus = $derived(health?.status || "unknown");
  const overallVariant = $derived(statusClass(health?.status));
</script>

<Card>
  <div class="health-grid-container">
    <div class="health-grid-header">
      <div>
        <p class="eyebrow">System Health</p>
        <p class="muted">Component status overview</p>
      </div>
      <Badge variant={overallVariant}>{overallStatus}</Badge>
    </div>

    {#if components.length === 0}
      <p class="muted">No component information available.</p>
    {:else}
      <div class="health-grid">
        {#each components as component}
          <div class="health-grid-item">
            <div class="health-grid-item__header">
              <span
                class={cn("status-dot", `status-dot--${statusClass(component.status)}`)}
                aria-label={component.status}
              ></span>
              <span class="health-grid-item__label">{component.label}</span>
            </div>
            <div class="health-grid-item__details">
              {#if component.responseTime}
                <span class="health-grid-item__time">{component.responseTime}</span>
              {:else if component.detail}
                <span class="health-grid-item__detail">{component.detail}</span>
              {/if}
              <Badge variant={statusClass(component.status)} class="health-grid-item__badge">
                {component.status}
              </Badge>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</Card>

<style>
  .health-grid-container {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .health-grid-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
  }

  .health-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .health-grid-item {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid #e4e4e7;
    border-radius: 8px;
    background: #fafafa;
    transition:
      border-color 0.2s ease,
      background-color 0.2s ease;
  }

  .health-grid-item:hover {
    border-color: #d4d4d8;
    background: #f4f4f5;
  }

  .health-grid-item__header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-dot--ok {
    background: #16a34a;
  }

  .status-dot--warn {
    background: #eab308;
  }

  .status-dot--error {
    background: #dc2626;
  }

  .status-dot--idle {
    background: #a1a1aa;
  }

  .health-grid-item__label {
    font-weight: 600;
    font-size: 0.875rem;
    color: #18181b;
  }

  .health-grid-item__details {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .health-grid-item__time {
    font-size: 0.75rem;
    color: #52525b;
    font-family: monospace;
  }

  .health-grid-item__detail {
    font-size: 0.75rem;
    color: #71717a;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .health-grid-item__badge {
    flex-shrink: 0;
  }

  @media (max-width: 768px) {
    .health-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
