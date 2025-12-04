<script lang="ts">
  import Card from "./ui/Card.svelte";

  interface MetricCardProps {
    title: string;
    value?: string | number;
    hint?: string;
  }

  let { title, value = "-", hint }: MetricCardProps = $props();
</script>

<Card>
  <div class="metric-card">
    <p class="eyebrow">{title}</p>
    <p class="metric-value">{value}</p>
    {#if hint}
      <p class="muted">{hint}</p>
    {/if}
  </div>
</Card>
