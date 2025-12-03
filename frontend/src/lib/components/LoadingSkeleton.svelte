<script lang="ts">
  export let variant: "metric" | "message" | "health" = "metric";
</script>

{#if variant === "metric"}
  <div class="card metric-card">
    <div class="skeleton skeleton--eyebrow"></div>
    <div class="skeleton skeleton--value"></div>
    <div class="skeleton skeleton--hint"></div>
  </div>
{:else if variant === "message"}
  <div class="card">
    <div class="card__header">
      <div>
        <div class="skeleton skeleton--eyebrow"></div>
        <div class="skeleton skeleton--muted"></div>
      </div>
      <div class="skeleton skeleton--pill"></div>
    </div>
    <ul class="message-list">
      {#each Array(3) as _}
        <li class="message">
          <div class="message-meta">
            <div class="skeleton skeleton--pill-small"></div>
            <div class="skeleton skeleton--timestamp"></div>
          </div>
          <div class="skeleton skeleton--message-body"></div>
          <div class="message-foot">
            <div class="skeleton skeleton--route"></div>
            <div class="skeleton skeleton--flight"></div>
            <div class="skeleton skeleton--pill-small"></div>
          </div>
        </li>
      {/each}
    </ul>
  </div>
{:else if variant === "health"}
  <div class="card">
    <div class="card__header">
      <div>
        <div class="skeleton skeleton--eyebrow"></div>
        <div class="skeleton skeleton--muted"></div>
      </div>
      <div class="skeleton skeleton--pill"></div>
    </div>
    <ul class="health-list">
      {#each Array(4) as _}
        <li class="health-row">
          <div class="skeleton skeleton--dot"></div>
          <div>
            <div class="skeleton skeleton--label"></div>
            <div class="skeleton skeleton--muted"></div>
          </div>
        </li>
      {/each}
    </ul>
  </div>
{/if}

<style>
  .skeleton {
    background: linear-gradient(90deg, #f1f5f9 25%, #e2e8f0 50%, #f1f5f9 75%);
    background-size: 200% 100%;
    animation: loading 1.5s ease-in-out infinite;
    border-radius: 4px;
  }

  @keyframes loading {
    0%,
    100% {
      background-position: 200% 0;
    }
    50% {
      background-position: -200% 0;
    }
  }

  .skeleton--eyebrow {
    width: 80px;
    height: 12px;
    margin-bottom: 0.5rem;
  }

  .skeleton--value {
    width: 120px;
    height: 32px;
    margin: 0.5rem 0;
  }

  .skeleton--hint {
    width: 100px;
    height: 14px;
    margin-top: 0.5rem;
  }

  .skeleton--muted {
    width: 150px;
    height: 14px;
    margin-top: 0.25rem;
  }

  .skeleton--pill {
    width: 60px;
    height: 24px;
    border-radius: 999px;
  }

  .skeleton--pill-small {
    width: 80px;
    height: 20px;
    border-radius: 999px;
  }

  .skeleton--timestamp {
    width: 120px;
    height: 14px;
  }

  .skeleton--message-body {
    width: 100%;
    height: 20px;
    margin: 0.5rem 0;
  }

  .skeleton--route {
    width: 100px;
    height: 14px;
  }

  .skeleton--flight {
    width: 80px;
    height: 14px;
  }

  .skeleton--label {
    width: 100px;
    height: 16px;
    margin-bottom: 0.25rem;
  }

  .skeleton--dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
  }
</style>

