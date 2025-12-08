<script lang="ts">
  import TypeDistributionChart from "../charts/TypeDistributionChart.svelte";
  import MessageTrendChart from "../charts/MessageTrendChart.svelte";
  import type { TrendDataPoint } from "../charts/MessageTrendChart.svelte";

  interface VisualizationSectionProps {
    byType: Record<string, number>;
    trendData: TrendDataPoint[];
  }

  let { byType, trendData }: VisualizationSectionProps = $props();
</script>

<svelte:boundary>
  <section class="grid grid--two">
    <TypeDistributionChart data={byType} title="Message Type Distribution" />
    <MessageTrendChart title="Message Trend" data={trendData} />
  </section>
</svelte:boundary>

<style>
  .grid {
    display: grid;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .grid--two {
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  }

  @media (max-width: 1024px) {
    .grid--two {
      grid-template-columns: 1fr;
    }
  }
</style>

