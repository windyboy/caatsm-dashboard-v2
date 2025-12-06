<script lang="ts">
  import Card from "./ui/Card.svelte";
  import Badge from "./ui/Badge.svelte";

  interface HealthCardProps {
    status?: string;
    items?: { label: string; status?: string; detail?: string }[];
  }

  let { status = "unknown", items = [] }: HealthCardProps = $props();

  const statusClass = (value?: string): "ok" | "warn" | "error" | "idle" => {
    if (!value) return "idle";
    const normalized = value.toLowerCase();
    if (["ok", "healthy", "connected"].includes(normalized)) return "ok";
    if (["degraded", "warning"].includes(normalized)) return "warn";
    if (["unhealthy", "error"].includes(normalized)) return "error";
    if (["not_configured"].includes(normalized)) return "idle";
    return "idle";
  };

  const statusVariant = $derived(statusClass(status));
</script>

<Card>
  <div class="card__header">
    <div>
      <p class="eyebrow">System Health</p>
      <p class="muted">Latest snapshot from the backend</p>
    </div>
    <Badge variant={statusVariant}>{status || "unknown"}</Badge>
  </div>

  {#if items.length === 0}
    <p class="muted">No component details available.</p>
  {:else}
    <ul class="health-list" role="list">
      {#each items as item (item.label)}
        <li class="health-row" role="listitem">
          <span class="status-dot {statusClass(item.status)}" aria-label={item.status || "unknown"}
          ></span>
          <div>
            <p class="health-label">{item.label}</p>
            <p class="muted">{item.detail || item.status || "No detail"}</p>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</Card>
