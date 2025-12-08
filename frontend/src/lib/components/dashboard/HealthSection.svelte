<script lang="ts">
  import HealthStatusGrid from "../HealthStatusGrid.svelte";
  import type { HealthSnapshot } from "$lib/types";

  interface HealthSectionProps {
    health: HealthSnapshot | null;
  }

  let { health }: HealthSectionProps = $props();
</script>

<section class="health-section">
  <HealthStatusGrid {health} />
</section>

<style>
  .health-section {
    margin-bottom: 1.5rem;
  }
</style>

