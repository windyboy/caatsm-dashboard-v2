<script lang="ts">
  export let status: string | undefined = "unknown";
  export let items: { label: string; status?: string; detail?: string }[] = [];

  const statusClass = (value?: string): string => {
    if (!value) return "pill--idle";
    const normalized = value.toLowerCase();
    if (["ok", "healthy", "connected"].includes(normalized)) return "pill--ok";
    if (["degraded", "warning"].includes(normalized)) return "pill--warn";
    if (["unhealthy", "error"].includes(normalized)) return "pill--error";
    return "pill--idle";
  };
</script>

<div class="card">
  <div class="card__header">
    <div>
      <p class="eyebrow">System Health</p>
      <p class="muted">Latest snapshot from the backend</p>
    </div>
    <span class="pill {statusClass(status)}">{status || "unknown"}</span>
  </div>

  {#if items.length === 0}
    <p class="muted">No component details available.</p>
  {:else}
    <ul class="health-list">
      {#each items as item}
        <li class="health-row">
          <span class="status-dot {statusClass(item.status)}"></span>
          <div>
            <p class="health-label">{item.label}</p>
            <p class="muted">{item.detail || item.status || "No detail"}</p>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>
