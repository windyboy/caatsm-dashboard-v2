<script lang="ts">
  import Card from "./ui/Card.svelte";

  interface TrendData {
    value: number; // percentage change
    label?: string;
  }

  interface EnhancedMetricCardProps {
    title: string;
    value?: string | number;
    hint?: string;
    trend?: TrendData;
    icon?: string;
    variant?: "default" | "primary";
  }

  let {
    title,
    value = "-",
    hint,
    trend,
    icon,
    variant = "default",
  }: EnhancedMetricCardProps = $props();

  const trendClass = $derived(trend ? (trend.value >= 0 ? "trend-up" : "trend-down") : "");
  const trendIcon = $derived(trend ? (trend.value >= 0 ? "↑" : "↓") : "");
</script>

<Card>
  <div class="enhanced-metric-card" class:enhanced-metric-card--primary={variant === "primary"}>
    <div class="metric-header">
      <p class="eyebrow">{title}</p>
      {#if icon}
        <span class="metric-icon" aria-hidden="true">{icon}</span>
      {/if}
    </div>
    <p class="metric-value">{value}</p>
    <div class="metric-footer">
      {#if trend}
        <span class="trend {trendClass}">
          <span class="trend-icon">{trendIcon}</span>
          <span>{Math.abs(trend.value).toFixed(1)}%</span>
        </span>
        {#if trend.label}
          <span class="muted trend-label">{trend.label}</span>
        {/if}
      {:else if hint}
        <span class="muted">{hint}</span>
      {/if}
    </div>
  </div>
</Card>

<style>
  .enhanced-metric-card {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-height: 120px;
  }

  .metric-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .metric-icon {
    font-size: 1.25rem;
    opacity: 0.6;
    flex-shrink: 0;
  }

  .metric-value {
    font-size: 2rem;
    font-weight: 800;
    margin: 0.1rem 0;
    letter-spacing: -0.02em;
    color: #18181b;
    transition: transform 0.3s ease-out;
  }

  .enhanced-metric-card:hover .metric-value {
    transform: scale(1.02);
  }

  .metric-footer {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-top: auto;
  }

  .trend {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-weight: 600;
    font-size: 0.875rem;
    transition: color 0.2s ease;
  }

  .trend-up {
    color: #16a34a;
  }

  .trend-down {
    color: #dc2626;
  }

  .trend-icon {
    font-size: 1rem;
    line-height: 1;
  }

  .trend-label {
    font-size: 0.75rem;
  }

  .enhanced-metric-card--primary .metric-value {
    color: #2563eb;
  }
</style>
