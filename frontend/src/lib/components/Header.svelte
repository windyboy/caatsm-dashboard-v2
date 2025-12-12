<script lang="ts">
  import * as NavigationMenu from "$lib/components/ui/navigation-menu";
  import { cn } from "$lib/utils";
  import { page } from "$app/stores";
  import { _ } from "svelte-i18n";
  import type { HTMLAttributes } from "svelte/elements";

  type ListItemProps = HTMLAttributes<HTMLAnchorElement> & {
    title: string;
    href: string;
    content: string;
  };
</script>

{#snippet ListItem({ title, content, href, class: className, ...restProps }: ListItemProps)}
  <li>
    <NavigationMenu.Link>
      {#snippet child()}
        <a
          {href}
          class={cn(
            "hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground block select-none space-y-1 rounded-md p-3 leading-none no-underline outline-none transition-colors",
            className
          )}
          {...restProps}
        >
          <div class="text-sm font-medium leading-none">{title}</div>
          <p class="text-muted-foreground line-clamp-2 text-sm leading-snug">
            {content}
          </p>
        </a>
      {/snippet}
    </NavigationMenu.Link>
  </li>
{/snippet}

<header class="topbar">
  <div class="topbar__brand">
    <span class="brand-mark" aria-hidden="true"></span>
    <div>
      <p class="eyebrow">CAATSM</p>
      <p class="title">{$_("header.title")}</p>
    </div>
  </div>

  <NavigationMenu.Root viewport={false} aria-label="Main navigation">
    <NavigationMenu.List>
      <NavigationMenu.Item>
        <NavigationMenu.Trigger>{$_("dashboard.title")}</NavigationMenu.Trigger>
        <NavigationMenu.Content>
          <ul class="grid gap-2 p-2 md:w-[400px] lg:w-[500px] lg:grid-cols-[.75fr_1fr]">
            <li class="row-span-3">
              <NavigationMenu.Link
                class="from-muted/50 to-muted bg-gradient-to-b outline-hidden flex h-full w-full select-none flex-col justify-end rounded-md p-6 no-underline focus:shadow-md"
              >
                {#snippet child({ props })}
                  <a {...props} href="/">
                    <div class="mb-2 mt-4 text-lg font-medium">CAATSM Dashboard</div>
                    <p class="text-muted-foreground text-sm leading-tight">
                      Real-time aviation telegram traffic monitoring and analysis.
                    </p>
                  </a>
                {/snippet}
              </NavigationMenu.Link>
            </li>
            {@render ListItem({
              href: "/",
              title: $_("dashboard.title"),
              content: "Main operational dashboard with real-time metrics and charts.",
            })}
            {@render ListItem({
              href: "/search",
              title: $_("search.title"),
              content: "Search and filter aviation telegrams with full-text search.",
            })}
          </ul>
        </NavigationMenu.Content>
      </NavigationMenu.Item>

      <NavigationMenu.Item>
        <NavigationMenu.Link>
          {#snippet child()}
            <a
              href="/"
              class="nav-link"
              aria-current={$page.url.pathname === "/" ? "page" : undefined}
            >
              {$_("dashboard.title")}
            </a>
          {/snippet}
        </NavigationMenu.Link>
      </NavigationMenu.Item>

      <NavigationMenu.Item>
        <NavigationMenu.Link>
          {#snippet child()}
            <a
              href="/search"
              class="nav-link"
              aria-current={$page.url.pathname === "/search" ? "page" : undefined}
            >
              {$_("search.title")}
            </a>
          {/snippet}
        </NavigationMenu.Link>
      </NavigationMenu.Item>
    </NavigationMenu.List>
  </NavigationMenu.Root>
</header>

<style>
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.5rem;
    background: #fff;
    border-bottom: 1px solid #e4e4e7;
    box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
    position: sticky;
    top: 0;
    z-index: 50;
  }

  .topbar__brand {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .brand-mark {
    width: 32px;
    height: 32px;
    background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
    border-radius: 6px;
    flex-shrink: 0;
  }

  .topbar__brand > div {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .nav-link {
    display: inline-flex;
    align-items: center;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #52525b;
    text-decoration: none;
    border-radius: 6px;
    transition: all 0.2s ease;
  }

  .nav-link:hover {
    background: #f4f4f5;
    color: #18181b;
  }

  .nav-link[aria-current="page"] {
    background: #eff6ff;
    color: #2563eb;
    font-weight: 600;
  }

  @media (max-width: 768px) {
    .topbar {
      padding: 0.75rem 1rem;
    }

    .topbar__brand > div {
      gap: 0.125rem;
    }

    .brand-mark {
      width: 28px;
      height: 28px;
    }
  }
</style>
