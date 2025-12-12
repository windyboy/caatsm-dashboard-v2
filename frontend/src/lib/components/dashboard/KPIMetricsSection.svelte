<script lang="ts">
  import EnhancedMetricCard from "../EnhancedMetricCard.svelte";

  interface KPIMetricsSectionProps {
    totalMessages: number;
    messagesPerSec: number;
    activeRoutes: number;
    systemStatus: string;
  }

  let { totalMessages, messagesPerSec, activeRoutes, systemStatus }: KPIMetricsSectionProps =
    $props();
</script>

<section class="grid grid--four">
  <EnhancedMetricCard
    title="Total Messages"
    value={totalMessages.toLocaleString()}
    hint="Last 24 hours"
  />
  <EnhancedMetricCard title="Messages/sec" value={messagesPerSec.toFixed(1)} hint="Last 1 min" />
  <EnhancedMetricCard
    title="Active Routes"
    value={activeRoutes.toLocaleString()}
    hint="Unique source→destination pairs"
  />
  <EnhancedMetricCard
    title="System Health"
    value={systemStatus.toUpperCase()}
    hint="Overall status"
  />
</section>

<style>
  .grid {
    display: grid;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .grid--four {
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  }

  @media (max-width: 1024px) {
    .grid--four {
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    }
  }

  @media (max-width: 768px) {
    .grid--four {
      grid-template-columns: 1fr;
    }
  }
</style>
