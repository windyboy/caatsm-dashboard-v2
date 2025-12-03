<script lang="ts">
  import { goto } from "$app/navigation";

  function goHome() {
    goto("/");
  }
</script>

<main class="page page--center">
  <div class="card card--compact">
    <p class="eyebrow">Login placeholder</p>
    <h1>Authentication to be wired up</h1>
    <p class="muted">
      The backend should handle authentication. Wire your preferred auth flow here when ready.
    </p>
    <div class="search-actions">
      <button class="button" type="button" onclick={goHome}>Back to dashboard</button>
    </div>
  </div>
</main>
