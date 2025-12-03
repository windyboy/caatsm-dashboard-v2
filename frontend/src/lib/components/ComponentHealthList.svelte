<!--
  @component ComponentHealthList
  Displays a list of component health statuses.
  
  @param {Array<{name: string, key: string, health: ComponentHealth}>} components - List of components to display
-->

<script lang="ts">
  import type { ComponentHealth } from "$lib/services/api";
  import { getStatusColor, getStatusIcon, getStatusLabel } from "$lib/utils/health";

  interface Props {
    components: Array<{ name: string; key: string; health?: ComponentHealth }>;
  }

  let { components }: Props = $props();
</script>

<div class="component-list">
  {#each components as { name, key, health }}
    {@const componentHealth = health || { status: "not_configured" } as ComponentHealth}
    <div class="component-item">
      <div class="component-name">
        <span class="component-icon {getStatusColor(componentHealth.status)}">
          {getStatusIcon(componentHealth.status)}
        </span>
        <span class="component-label">{name}</span>
      </div>
      <div class="component-status">
        <span class="status-text {getStatusColor(componentHealth.status)}">
          {getStatusLabel(componentHealth.status)}
        </span>
        {#if componentHealth.response_time}
          <span class="response-time" title="Response time">
            {componentHealth.response_time}
          </span>
        {/if}
        {#if componentHealth.message}
          <span class="component-message" title={componentHealth.message}>
            ⓘ
          </span>
        {/if}
      </div>
    </div>
  {/each}
</div>

<style>
  .component-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .component-item {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: center;
    gap: 1rem;
    padding: 0.5rem;
    border-radius: 0.375rem;
    background: rgba(248, 250, 252, 0.5);
  }

  .component-name {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 0;
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
    gap: 0.5rem;
  }

  .status-text {
    font-size: 0.75rem;
    font-weight: 500;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    border: 1px solid;
  }

  .response-time {
    font-size: 0.7rem;
    color: #64748b;
    font-family: monospace;
  }

  .component-message {
    font-size: 0.75rem;
    color: #64748b;
    cursor: help;
  }
</style>

