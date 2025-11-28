# CAATSM Dashboard Frontend

SvelteKit frontend for the CAATSM Dashboard, built with **Deno 2.0+** (recommended) or **Node.js 20+**.

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

````bash
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

The frontend includes end-to-end tests using Playwright to ensure functionality across different browsers and scenarios.

### Running Tests

```bash
# Run all E2E tests (using Taskfile/Makefile - recommended)
task frontend:test
# or
make frontend-test

# Or manually with Deno
deno task test

# Or manually with Node.js
npm run test

# Run tests in headed mode (visible browser)
npx playwright test --headed

# Run specific test file
npx playwright test tests/api.test.ts

# Generate test report
npx playwright show-report
````

### Test Configuration

- **Framework:** Playwright
- **Browsers:** Chromium, Firefox, WebKit (Safari)
- **Test Files:** `tests/*.spec.ts`
- **Configuration:** `playwright.config.ts`

### Writing Tests

Tests are located in the `tests/` directory and follow Playwright's test structure:

```typescript
import { expect, test } from "@playwright/test";

test("search functionality", async ({ page }) => {
  await page.goto("/");
  // Test implementation
});
```

### Test Coverage

Current E2E tests cover:

- Basic page loading
- Search form interactions
- Real-time message streaming
- Statistics display
- WebSocket connections

````
## Unit Testing

The frontend includes unit tests using Vitest to test business logic, stores, and services.

### Running Unit Tests

```bash
# Run all unit tests (using Taskfile/Makefile - recommended)
task frontend:test:unit
# or
make frontend-test-unit

# Or manually with Deno
deno task test:unit

# Or manually with Node.js
npm run test:unit

# Run tests in watch mode
npm run test:unit -- --watch

# Run tests with UI
npm run test:unit:ui

# Run tests with coverage
npm run test:unit:coverage
````

### Test Configuration

- **Framework:** Vitest
- **Environment:** jsdom (for DOM APIs)
- **Test Files:** `tests/unit/*.test.ts`
- **Configuration:** `vitest.config.ts`
- **Setup:** `tests/setup.ts` (mocks WebSocket and logger)

### Writing Unit Tests

Unit tests are located in the `tests/unit/` directory:

```typescript
import { describe, expect, it } from "vitest";
import { get } from "svelte/store";
import { messages } from "../../src/lib/stores/messages";

describe("messages store", () => {
  it("should add a message", () => {
    const telegram = { message_id: "TEST-001" /* ... */ };
    messages.add(telegram);
    expect(get(messages)).toHaveLength(1);
  });
});
```

### Test Coverage

Current unit tests cover:

- WebSocketClient: Connection, reconnection, message handling
- Stores: Messages store (add, limit enforcement, clear)
- Stores: Stats store (setTotal, setByPriority, setByType, reset)
- Error handling: Invalid JSON, connection errors
- Business logic: Message limiting (MAX_MESSAGES = 50), exponential backoff

````
## Building

### With Deno

```bash
deno task build
````

### With Node.js

```bash
npm run build
```

The built files will be in the `build/` directory. The Go backend can serve these static files in production.

## Deployment

### Production Build

1. **Build the frontend:**

   ```bash
   cd frontend
   deno task build  # or npm run build
   ```

2. **Backend serves static files:**
   The Go backend automatically serves the built frontend from `frontend/build/` when running in production mode.

3. **Environment variables:**
   Set `VITE_API_BASE_URL` and `VITE_WS_URL` to match your backend deployment URLs.

### Development vs Production

- **Development:** Hot reload, source maps, detailed error messages
- **Production:** Minified code, optimized assets, error boundaries

### Static Asset Optimization

- **UnoCSS:** Atomic CSS generation for minimal bundle size
- **Vite:** Fast builds with tree-shaking and code splitting
- **Svelte:** Compile-time optimizations and small runtime

### CDN Deployment

The built files are static and can be deployed to any CDN or static hosting service, with the API endpoints configured via environment variables.

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

## Components

### MessageItem.svelte

Displays individual telegram messages with a modern card design featuring:

- **Message ID** and flight number
- **Timestamp** formatted as local time
- **Message content** with proper text wrapping
- **Type tags** (AFTN, SITA, ACARS, CPDLC) with color coding
- **Priority level** indicator
- **Route information** (source → destination) when available

**Props:**

- `telegram: Telegram` - The telegram data object

**Styling:** Glassmorphism effects with hover animations and responsive design.

### SearchForm.svelte

Interactive search form component with autocomplete functionality:

- **Query input** with real-time autocomplete suggestions (debounced 500ms)
- **Type filter** dropdown (AFTN, SITA, ACARS, CPDLC)
- **Priority filter** dropdown (Urgent, Operational, Routine)
- **Time range** input (currently disabled, placeholder for future feature)
- **Search and Reset buttons**

**Events:**

- `search` - Dispatched with search parameters
- `reset` - Dispatched when reset button is clicked

**Features:** Autocomplete API integration, form validation, keyboard navigation.

#### Autocomplete Functionality

The query input provides intelligent search suggestions as users type:

- **Trigger**: Activates after typing 2+ characters
- **Debouncing**: 500ms delay to prevent excessive API calls
- **API Endpoint**: `GET /api/autocomplete?term={query}&size=5`
- **Suggestions**: Up to 5 results from Meilisearch (flight numbers, message content, IDs)
- **Interaction**: Click suggestion to auto-fill query and trigger search
- **Error Handling**: Gracefully handles API failures without disrupting UX
- **Performance**: Cancels pending requests on new input

### SearchResults.svelte

Displays paginated search results:

- **Result count** display with formatted numbers
- **Message list** using MessageItem components
- **Empty state** with helpful messaging when no results found

**Props:**

- `telegrams: Telegram[]` - Array of telegram results
- `total: number` - Total number of matching results

### StatsCard.svelte

Flexible statistics display component supporting multiple data types:

- **Total count** - Single number display
- **Priority breakdown** - List of counts by priority level
- **Type breakdown** - List of counts by message type

**Props:**

- `title: string` - Card title
- `type: "total" | "priority" | "type"` - Display mode

**Reactive:** Automatically updates from the stats store.

### LiveStream.svelte

Real-time telegram streaming component:

- **WebSocket connection** for live updates
- **Auto-scroll** to newest messages
- **Real-time indicator** with animated pulse
- **Message display** using MessageItem components
- **Empty state** when waiting for messages

**Features:** Automatic reconnection, smooth scrolling, connection status display.

## API Integration

The frontend communicates with the Go backend through REST API calls and WebSocket connections for real-time updates.

### REST API Endpoints

- `GET /api/search` - Search telegrams with query parameters
- `POST /api/search` - Advanced search with JSON payload
- `GET /api/autocomplete?term=...&size=5` - Get autocomplete suggestions
- `GET /api/stats/total` - Get total message count (24h)
- `GET /api/stats/priority` - Get priority breakdown
- `GET /api/stats/type` - Get message type breakdown

### WebSocket Connection

- **Endpoint:** `GET /ws`
- **Protocol:** JSON messages with type-based routing
- **Message Types:**
  - `message` - New telegram received
  - `stats-total` - Updated total count
  - `stats-priority` - Updated priority stats
  - `stats-type` - Updated type stats
- **Features:** Automatic reconnection, heartbeat ping/pong

### Services

#### api.ts

Handles REST API communication:

- `search(params)` - Execute search queries
- `autocomplete(term, size)` - Get suggestions
- `getStats()` - Fetch statistics

#### websocket.ts

Manages WebSocket connection:

- Connection lifecycle management
- Message parsing and routing
- Reconnection logic
- Error handling

### Stores

#### messages.ts

Manages live telegram stream:

- Reactive message array
- Auto-cleanup of old messages
- WebSocket integration

#### stats.ts

Handles statistics data:

- Total, priority, and type breakdowns
- Real-time updates via WebSocket
- Reactive state management

#### websocket.ts

WebSocket connection state:

- Connection status
- Reconnection attempts
- Error states

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
