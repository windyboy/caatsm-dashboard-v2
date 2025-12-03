<!--
  @component HealthStatusIndicator
  Displays a status badge with icon and label.
  
  @param {string} status - Status value (ok, healthy, error, etc.)
  @param {boolean} [showIcon=true] - Whether to show the status icon
  @param {boolean} [showLabel=true] - Whether to show the status label
-->

<script lang="ts">
  import { getStatusColor, getStatusIcon, getStatusLabel } from "$lib/utils/health";

  interface Props {
    status: string;
    showIcon?: boolean;
    showLabel?: boolean;
  }

  let { status, showIcon = true, showLabel = true }: Props = $props();
</script>

<span class="status-badge {getStatusColor(status)}">
  {#if showIcon}{getStatusIcon(status)}{/if}
  {#if showLabel}{getStatusLabel(status)}{/if}
</span>

<style>
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
</style>

