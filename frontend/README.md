# CAATSM Dashboard Frontend

The frontend is a SvelteKit app that runs on top of Deno 2 (recommended) or Node.js 20. It talks to the Go backend at http://localhost:3002 and streams live data over WebSocket.

---

## Prerequisites

- Deno 2.x (preferred) or Node.js 20+
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
     - `deno task dev`
     - `npm run dev` (if using Node.js)

3. Open http://localhost:5173 in your browser.

The dev server proxies API calls and WebSocket traffic to http://localhost:3002.

---

## Core Scripts

| Task | Deno | Node.js |
| --- | --- | --- |
| Start dev server | `deno task dev` | `npm run dev` |
| Build for production | `deno task build` | `npm run build` |
| Preview build | `deno task preview` | `npm run preview` |
| Type check | `deno task check` | `npm run check` |
| Format | `deno fmt` | `npm run format` |
| Lint | `deno lint` | `npm run lint` |

---

## Testing

- Unit tests (Vitest): `deno task test:unit` or `npm run test:unit`
- E2E tests (Playwright): `deno task test` or `npm run test`

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

The frontend uses the following key dependencies:
- **@sveltejs/kit**: `^2.49.0` - Latest stable SvelteKit 2.x
- **@sveltejs/vite-plugin-svelte**: `^4.0.4` - Latest vite-plugin-svelte 4.x (compatible with vite 5.x)
- **vite**: `^5.4.21` - Latest vite 5.x

**Version Strategy**: We stay on vite 5.x and vite-plugin-svelte 4.x for stability. SvelteKit 2.49.0 supports vite 6/7 and vite-plugin-svelte 5/6, but upgrading to these major versions requires testing for breaking changes. The `deno.json` also pins to major versions 4 and 5 for consistency.

---

## Troubleshooting

- **Dev server cannot reach backend**: ensure the backend listens on port 3002.
- **WebSocket fails**: confirm `VITE_WS_URL` matches backend host and that CORS/origin rules allow requests.
- **Type errors on fresh clone**: run `deno cache --reload` or `npm install` before starting dev.
- **Live data missing**: run the sync worker and publish sample messages (`task publish-stream:fast`).

Keep the backend running alongside the frontend for a complete experience.