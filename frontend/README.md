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

## Troubleshooting

- **Dev server cannot reach backend**: ensure the backend listens on port 3002.
- **WebSocket fails**: confirm `VITE_WS_URL` matches backend host and that CORS/origin rules allow requests.
- **Type errors on fresh clone**: run `deno cache --reload` or `npm install` before starting dev.
- **Live data missing**: run the sync worker and publish sample messages (`task publish-stream:fast`).

Keep the backend running alongside the frontend for a complete experience.