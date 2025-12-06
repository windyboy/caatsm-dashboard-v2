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

The frontend uses **shadcn-svelte** for accessible, beautifully-designed UI components. Components are copied into the project (not installed as dependencies), allowing full customization.

```svelte
<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
</script>

<Button variant="default" size="default">Click me</Button>
```

### Component Structure

- **Base components** (`lib/components/ui/`): Reusable UI primitives from shadcn-svelte
  - `Button.svelte` - Accessible button component (wrapper around shadcn-svelte Button)
  - `Input.svelte` - Accessible input component with label support
  - `Card.svelte` - Card container component
  - `Badge.svelte` - Badge/pill component with custom variant mapping
  - `Select.svelte` - Select dropdown component

- **Feature components** (`lib/components/`): Higher-level components using base UI components
  - `Header.svelte` - Application header
  - `MetricCard.svelte` - Metric display card
  - `HealthCard.svelte` - System health display
  - `MessageList.svelte` - Message list display
  - `LoadingSkeleton.svelte` - Loading state skeleton

### Styling

Components use **TailwindCSS** for styling with shadcn-svelte's design system:

- TailwindCSS utility classes for all styling
- CSS variables for theming (defined in `app.css`)
- Built-in accessibility features (ARIA attributes, keyboard navigation, focus management)
- Dark mode support via CSS variables

To add new shadcn-svelte components:

```bash
bun x shadcn-svelte@latest add <component-name>
```
