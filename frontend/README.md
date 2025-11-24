# CAATSM Dashboard Frontend

SvelteKit frontend for the CAATSM Dashboard, built with Node.js + npm.

## Development

### Prerequisites

You can use either **Deno** or **Node.js** for development:

- **Option 1: Deno 2.0+** (recommended for this project)
- **Option 2: Node.js 20+** with npm/pnpm/yarn

### Setup with Deno (Recommended)

```bash
# No installation needed! Deno downloads dependencies automatically
# Start development server
deno task dev
```

### Setup with Node.js

```bash
# Install dependencies
npm install

# Start development server
npm run dev
# or
deno task dev:npm
```

The frontend will be available at `http://localhost:5173` (or the port shown in the terminal).

The dev server is configured to proxy API requests to the Go backend at `http://localhost:3002`.

### Environment Variables

Create a `.env` file in the frontend directory (optional, defaults are provided):

```env
VITE_API_BASE_URL=http://localhost:3002
VITE_WS_URL=ws://localhost:3002/ws
```

## Available Scripts

### Using Deno (Recommended)

```bash
# Development
deno task dev        # Start development server with hot reload (uses Deno)

# Building
deno task build      # Build for production (uses Deno)

# Preview
deno task preview    # Preview production build locally (uses Deno)

# Code Quality
deno task check      # Type check with Deno
deno fmt             # Format code with Deno formatter
deno lint            # Lint code with Deno linter
```

### Using Node.js/npm

```bash
# Development
npm run dev          # Start development server with hot reload
# or
deno task dev:npm    # Run npm dev via Deno

# Building
npm run build        # Build for production
# or
deno task build:npm  # Run npm build via Deno

# Preview
npm run preview      # Preview production build locally
# or
deno task preview:npm # Run npm preview via Deno

# Code Quality
npm run check        # Type check with svelte-check
npm run check:watch  # Type check in watch mode
npm run format       # Format code with Prettier
npm run lint         # Lint code with Prettier

# Testing
npm run test         # Run Playwright E2E tests
npm run test:e2e     # Run Playwright E2E tests (same as test)
```

## Building

### With Deno

```bash
deno task build
```

### With Node.js

```bash
npm run build
```

The built files will be in the `build/` directory. The Go backend can serve these static files in production.

## Type Checking

### With Deno

```bash
deno task check
```

This runs `svelte-kit sync` and then `deno check` to validate types.

### With Node.js

```bash
npm run check
```

This runs `svelte-kit sync` to generate TypeScript definitions and then `svelte-check` to validate types.

## Formatting

### With Deno

```bash
deno fmt
```

Formats code using Deno's built-in formatter.

### With Node.js

```bash
npm run format
```

Formats all code using Prettier with Svelte plugin support.

## Project Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/     # Svelte components
│   │   │   ├── MessageItem.svelte
│   │   │   ├── SearchForm.svelte
│   │   │   ├── SearchResults.svelte
│   │   │   ├── StatsCard.svelte
│   │   │   └── LiveStream.svelte
│   │   ├── stores/         # Svelte stores (state management)
│   │   │   ├── messages.ts
│   │   │   ├── stats.ts
│   │   │   └── websocket.ts
│   │   ├── services/        # API and WebSocket clients
│   │   │   ├── api.ts
│   │   │   └── websocket.ts
│   │   └── utils/           # Utilities and types
│   │       └── types.ts
│   ├── routes/              # SvelteKit routes
│   │   ├── +page.svelte     # Dashboard page
│   │   ├── +layout.svelte   # Root layout
│   │   └── search/
│   │       └── +page.svelte # Search page
│   ├── app.html             # HTML template
│   ├── app.css              # Global styles (UnoCSS)
│   └── app.ts               # App entry point
├── static/                  # Static assets
├── tests/                   # Playwright E2E tests
├── package.json             # npm dependencies and scripts
├── vite.config.ts           # Vite configuration
├── svelte.config.js         # SvelteKit configuration
├── uno.config.ts            # UnoCSS configuration
└── tsconfig.json            # TypeScript configuration
```

## Features

- **Real-time updates** via WebSocket connection to Go backend
- **Search functionality** with autocomplete
- **Statistics dashboard** with live updates
- **Modern UI** with glassmorphism effects
- **UnoCSS** for atomic CSS styling (Tailwind-compatible)
- **TypeScript** for type safety
- **SvelteKit** for routing and SSR capabilities

## Technology Stack

- **Svelte 5** - Modern reactive framework
- **SvelteKit 2** - Full-stack Svelte framework
- **Vite 7** - Next-generation build tool
- **UnoCSS** - Atomic CSS engine (Tailwind-compatible)
- **TypeScript** - Type-safe JavaScript
- **Playwright** - E2E testing

## Development Workflow

1. **Start the Go backend** (in project root):
   ```bash
   make dev
   # or
   ./bin/caatsm -config config/config.local.toml
   ```

2. **Start the frontend** (in frontend directory):

   **With Deno (Recommended)**:
   ```bash
   deno task dev
   ```

   **With Node.js**:
   ```bash
   npm install  # First time only
   npm run dev
   ```

3. **Access the application**:
   - Frontend: `http://localhost:5173`
   - Backend API: `http://localhost:3002`
   - WebSocket: `ws://localhost:3002/ws`

## Deno vs Node.js

This project supports both Deno and Node.js for development:

### Using Deno (Recommended)

**Advantages**:
- ✅ No `npm install` needed - dependencies are downloaded automatically
- ✅ Built-in TypeScript support
- ✅ Built-in formatter and linter
- ✅ Better security model
- ✅ Faster startup time
- ✅ Matches the original plan's intent

**Usage**:
```bash
deno task dev      # Start dev server
deno task build   # Build for production
deno task check   # Type check
deno fmt          # Format code
deno lint         # Lint code
```

### Using Node.js

**Advantages**:
- ✅ More familiar to most developers
- ✅ Larger ecosystem
- ✅ Better IDE support in some cases

**Usage**:
```bash
npm install        # Install dependencies
npm run dev        # Start dev server
npm run build      # Build for production
npm run check      # Type check
npm run format     # Format code
```

**Recommendation**: Use **Deno** for this project as it aligns with the original plan and provides a better development experience with zero configuration.

