<script lang="ts">
  import { getWebSocketStore } from "$lib/stores/websocket.svelte";
  import Select, { type SelectOption } from "./ui/Select.svelte";
  import Badge from "./ui/Badge.svelte";
  import { Button } from "./ui/button/index.js";
  import { cn } from "$lib/utils.js";

  interface TopControlBarProps {
    timeWindow?: string;
    onTimeWindowChange?: (value: string) => void;
    onRefresh?: () => void;
  }

  let { timeWindow = $bindable("last_24h"), onTimeWindowChange, onRefresh }: TopControlBarProps = $props();

  // Initialize store using $effect for reactive initialization
  // The store is a singleton, so it's safe to initialize directly
  const wsStore = getWebSocketStore();

  const timeWindowOptions: SelectOption[] = [
    { value: "last_1h", label: "Last 1 hour" },
    { value: "last_24h", label: "Last 24 hours" },
    { value: "last_7d", label: "Last 7 days" },
    { value: "custom", label: "Custom range" },
  ];

  const connectionStatus = $derived(wsStore?.status ?? "disconnected");
  
  // Use simple derived for direct mapping
  const statusVariant = $derived(
    connectionStatus === "connected" ? "ok" :
    connectionStatus === "connecting" ? "warn" :
    (connectionStatus === "disconnected" || connectionStatus === "error") ? "error" :
    "idle"
  );

  const statusLabel = $derived(
    connectionStatus === "connected" ? "Connected" :
    connectionStatus === "connecting" ? "Connecting..." :
    connectionStatus === "disconnected" ? "Disconnected" :
    connectionStatus === "error" ? "Error" :
    "Unknown"
  );

  function handleTimeWindowChange(value: string) {
    onTimeWindowChange?.(value);
  }

  function handleRefresh() {
    onRefresh?.();
  }
</script>

<div class="top-control-bar">
  <div class="top-control-bar__left">
    <div class="connection-status">
      <span class={cn("status-dot", `status-dot--${statusVariant}`)} aria-hidden="true"></span>
      <div aria-label={`Connection status: ${statusLabel}`}>
        <Badge variant={statusVariant}>
          {statusLabel}
        </Badge>
      </div>
    </div>
  </div>

  <div class="top-control-bar__center">
    <Select options={timeWindowOptions} bind:value={timeWindow} onValueChange={handleTimeWindowChange} />
  </div>

  <div class="top-control-bar__right">
    <Button variant="ghost" onclick={handleRefresh} type="button">
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        class="mr-1.5"
        aria-hidden="true"
      >
        <path
          d="M8 1.33334V4.66668M8 1.33334L5.33333 4.00001M8 1.33334L10.6667 4.00001M8 14.6667V11.3333M8 14.6667L10.6667 12M8 14.6667L5.33333 12M2.66667 8H6M10 8H13.3333M2.66667 8L4 6.66668M2.66667 8L4 9.33335M13.3333 8L12 6.66668M13.3333 8L12 9.33335"
          stroke="currentColor"
          stroke-width="1.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
      Refresh
    </Button>
  </div>
</div>

<style>
  .top-control-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 1rem 1.5rem;
    background: #fff;
    border-bottom: 1px solid #e4e4e7;
    flex-wrap: wrap;
  }

  .top-control-bar__left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .top-control-bar__center {
    flex: 1;
    min-width: 200px;
    max-width: 300px;
  }

  .top-control-bar__right {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .connection-status {
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

  .status-label {
    font-size: 0.875rem;
    color: #52525b;
    font-weight: 500;
  }

  @media (max-width: 768px) {
    .top-control-bar {
      flex-direction: column;
      align-items: stretch;
    }

    .top-control-bar__center {
      max-width: 100%;
    }

    .status-label {
      display: none;
    }
  }
</style>
