# CAATSM Dashboard Frontend

The frontend is a SvelteKit 2.x app with Svelte 5, Vite 7, and UnoCSS. It runs on Deno 2.0+ by default, with Bun as the fast fallback and Node.js 18+ (20+ recommended) supported. It talks to the Go backend at http://localhost:3002 and streams live data over WebSocket.

---

## Prerequisites

- Deno 2.x (preferred) or Bun (fallback) or Node.js 18+ (20+ recommended)
- Make or Task (optional shortcuts)
- Backend running on http://localhost:3002 (see project root README)

---

## Quick Start

1. From the project root, set up dependencies:
   - `make frontend-setup` (Deno)  
   - `task frontend:setup` (same as above)

2. Start the dev server:
   - `make frontend-dev` or `task frontend:dev`  
   - Or inside `frontend/`:
     - `deno task dev` (HMR disabled in Deno to avoid Vite 7 WebSocket issues)
     - `bun run dev` (full HMR)
     - `npm run dev`

3. Open http://localhost:5173 in your browser.

The dev server proxies API calls and WebSocket traffic to http://localhost:3002.

---

## Core Scripts

| Task | Deno | Bun | Node.js |
| --- | --- | --- | --- |
| Start dev server | `deno task dev` | `bun run dev` | `npm run dev` |
| Build for production | `deno task build` | `bun run build` | `npm run build` |
| Preview build | `deno task preview` | `bun run preview` | `npm run preview` |
| Type check | `deno task check` | `bun run check` | `npm run check` |
| Format | `deno fmt` | `bun run format` | `npm run format` |
| Lint | `deno lint` | `bun run lint` | `npm run lint` |

---

## Testing

- Unit tests (Vitest): `deno task test:unit` or `bun run test:unit` or `npm run test:unit`
- E2E tests (Playwright): `deno task test:e2e` (requires Bun for the Playwright runner) or `bun run test` or `npm run test:e2e`

---

## Environment Variables

Create `frontend/.env` (optional):

```
VITE_API_BASE_URL=http://localhost:3002
VITE_WS_URL=ws://localhost:3002/ws
```

These default values already target the local backend.

---

## Project Structure (partial)

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/      // Svelte UI pieces
│   │   ├── stores/          // Reactive app state
│   │   ├── services/        // REST + WebSocket clients
│   │   └── utils/           // Shared helpers
│   ├── routes/              // SvelteKit pages
│   └── app.html / app.css   // Root templates
├── static/                  // Public assets
├── tests/                   // Playwright specs
├── deno.json / package.json // Runtime config
└── vite.config.ts           // Build settings
```

---

## Dependency Versions

Key dependencies:
- **Svelte 5** - Runes-first reactivity (currently `^5.45.4`)
- **@sveltejs/kit**: `^2.49.1` - Latest SvelteKit 2.x
- **@sveltejs/vite-plugin-svelte**: `^6.2.1` - Plugin for Vite 7
- **vite**: `^7.2.6` - Bundler
- **unocss**: `^66.5.10` - Utility-first styling

**Version Strategy**: We track Vite 7 + vite-plugin-svelte 6 and keep SvelteKit on the latest 2.x line. `deno.json` pins the matching major versions for Deno tasks. Upgrading to future Vite or SvelteKit majors should include a compatibility pass for Svelte 5 runes and UnoCSS.

---

## Troubleshooting

- **Dev server cannot reach backend**: ensure the backend listens on port 3002.
- **WebSocket fails**: confirm `VITE_WS_URL` matches backend host and that `websocket.allowed_origins` in backend config includes the frontend origin.
- **Type errors on fresh clone**: run `deno cache --reload` or `bun install` before starting dev.
- **Live data missing**: ensure backend is running and publish sample messages (`task backend:publish-stream:fast`).
- **Deno dev HMR disabled**: when `deno task dev` sets `USE_DENO=true`, Vite 7 HMR is turned off to avoid WebSocket broken pipe errors. Use `bun run dev` for hot reload.

Keep the backend running alongside the frontend for a complete experience.
