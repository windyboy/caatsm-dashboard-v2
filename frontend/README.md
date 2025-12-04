# CAATSM Dashboard Frontend (Simplified)

A lean SvelteKit dashboard that surfaces message totals, a basic health snapshot, and the latest telegrams from the backend.

## Quick start

- Install deps with your preferred tool (npm/bun/deno).
- Run `npm run dev` (or `deno task dev`) and open http://localhost:5173.
- API base defaults to http://localhost:3002. Override with `VITE_API_BASE_URL` if needed.

## Available scripts

- `npm run dev` – start the dev server
- `npm run build` – production build
- `npm run preview` – preview the build
- `npm run check` – type check via `svelte-check`
- `npm run format` / `npm run lint` – run Prettier

## Pages

- `/` – dashboard with totals, system health, and latest messages
- `/search` – single-field search that hits `/api/search`
- `/login` – placeholder ready for your auth flow

Tests and auxiliary demo routes were removed to keep the frontend small and focused.

## Component Development

### UI Component Library

The frontend uses **@melt-ui/svelte** for accessible, headless UI components. All UI components follow the builder pattern:

```svelte
<script lang="ts">
  import { createButton, melt } from "@melt-ui/svelte";

  const { elements, states } = createButton({
    disabled: $derived(false),
  });
</script>

<button use:melt={$elements.root} class="button">
  <slot />
</button>
```

### Component Structure

- **Base components** (`lib/components/ui/`): Reusable UI primitives built with @melt-ui/svelte
  - `Button.svelte` - Accessible button component
  - `Input.svelte` - Accessible input component
  - `Card.svelte` - Card container component
  - `Badge.svelte` - Badge/pill component

- **Feature components** (`lib/components/`): Higher-level components using base UI components
  - `Header.svelte` - Application header
  - `MetricCard.svelte` - Metric display card
  - `HealthCard.svelte` - System health display
  - `MessageList.svelte` - Message list display
  - `LoadingSkeleton.svelte` - Loading state skeleton

### Styling

Components maintain existing styles from `app.css` while gaining built-in accessibility features:

- ARIA attributes automatically applied
- Keyboard navigation support
- Focus management
- Screen reader compatibility

All styling is controlled through UnoCSS and custom CSS classes, maintaining full design system control.
