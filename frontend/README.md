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
