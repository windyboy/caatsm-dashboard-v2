<script lang="ts">
  import { page } from "$app/stores";
</script>

<header class="topbar">
  <div class="topbar__brand">
    <span class="brand-mark"></span>
    <div>
      <p class="eyebrow">CAATSM</p>
      <p class="title">Operations Dashboard</p>
    </div>
  </div>

  <nav class="topbar__nav">
    <a class="nav-link" href="/" aria-current={$page.url.pathname === "/" ? "page" : undefined}>Dashboard</a>
    <a class="nav-link" href="/search" aria-current={$page.url.pathname === "/search" ? "page" : undefined}>Search</a>
  </nav>
</header>
