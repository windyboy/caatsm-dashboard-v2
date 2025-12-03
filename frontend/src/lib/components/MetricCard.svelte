<script lang="ts">
  export let title: string;
  export let value: string | number = "-";
  export let hint: string | undefined = undefined;
</script>

<div class="card metric-card">
  <p class="eyebrow">{title}</p>
  <p class="metric-value">{value}</p>
  {#if hint}
    <p class="muted">{hint}</p>
  {/if}
</div>
