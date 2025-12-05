/// <reference types="@sveltejs/kit" />
/// <reference lib="webworker" />

// Type assertion for service worker global scope
declare const self: ServiceWorkerGlobalScope;

// Service worker is only available in production builds
// In development, $service-worker module may not be available
// Use dynamic import with error handling
let build: string[] = [];
let files: string[] = [];
let version: string = "dev";

// Initialize service worker module (only available in production builds)
// This is async, so we need to handle it properly
const initServiceWorker = async (): Promise<void> => {
  try {
    const serviceWorker = await import("$service-worker");
    build = serviceWorker.build;
    files = serviceWorker.files;
    version = serviceWorker.version;
  } catch {
    // In development, $service-worker module is not available
    // This is expected and service worker will not be registered
    console.debug("Service worker module not available (development mode)");
  }
};

// Initialize immediately (top-level await not available in service worker context)
initServiceWorker();

const getCacheName = (): string => `cache-${version}`;
const getAssets = (): string[] => [...build, ...files];

// Helper to check if request should be skipped
const shouldSkipRequest = (request: Request): boolean => {
  const url = new URL(request.url);
  
  // Skip service worker script itself to avoid infinite loops
  if (url.pathname.includes("/service-worker.js")) {
    return true;
  }
  
  // Skip Vite HMR requests in development
  if (url.pathname.includes("/@vite/client") || url.pathname.includes("/@fs/")) {
    return true;
  }
  
  // Skip WebSocket connections
  if (url.protocol === "ws:" || url.protocol === "wss:") {
    return true;
  }
  
  // Skip API and WebSocket proxy paths
  if (url.pathname.startsWith("/api/") || url.pathname.startsWith("/ws")) {
    return true;
  }
  
  return false;
};

self.addEventListener("install", (event: ExtendableEvent) => {
  const assets = getAssets();
  
  // Only cache if we have assets
  if (assets.length === 0) {
    console.debug("Service worker: No assets to cache, skipping install");
    // Still skip waiting to activate immediately
    event.waitUntil(self.skipWaiting());
    return;
  }

  event.waitUntil(
    caches
      .open(getCacheName())
      .then((cache) => {
        console.debug(`Service worker: Caching ${assets.length} assets`);
        return cache.addAll(assets);
      })
      .then(() => {
        console.debug("Service worker: Installation complete");
        return self.skipWaiting();
      })
      .catch((error) => {
        console.error("Service worker: Installation failed", error);
        // Still skip waiting even if caching fails
        return self.skipWaiting();
      })
  );
});

self.addEventListener("activate", (event: ExtendableEvent) => {
  event.waitUntil(
    caches
      .keys()
      .then(async (keys) => {
        const cacheName = getCacheName();
        const deletePromises = keys
          .filter((key) => key !== cacheName)
          .map((key) => {
            console.debug(`Service worker: Deleting old cache ${key}`);
            return caches.delete(key);
          });
        await Promise.all(deletePromises);
        console.debug("Service worker: Activation complete");
        return self.clients.claim();
      })
      .catch((error) => {
        console.error("Service worker: Activation failed", error);
        // Still claim clients even if cache cleanup fails
        return self.clients.claim();
      })
  );
});

self.addEventListener("fetch", (event: FetchEvent) => {
  // Only handle GET requests
  if (event.request.method !== "GET") {
    return;
  }

  // Skip requests that should not be cached
  if (shouldSkipRequest(event.request)) {
    return;
  }

  event.respondWith(
    caches
      .match(event.request)
      .then((cachedResponse) => {
        if (cachedResponse) {
          return cachedResponse;
        }
        // Fallback to network
        return fetch(event.request).catch((error) => {
          console.error("Service worker: Fetch failed", error);
          // Return a basic error response
          return new Response("Network error", {
            status: 503,
            statusText: "Service Unavailable",
          });
        });
      })
      .catch((error) => {
        console.error("Service worker: Cache match failed", error);
        // Fallback to network
        return fetch(event.request).catch((fetchError) => {
          console.error("Service worker: Network fetch also failed", fetchError);
          return new Response("Network error", {
            status: 503,
            statusText: "Service Unavailable",
          });
        });
      })
  );
});

// Handle errors globally
self.addEventListener("error", (event: ErrorEvent) => {
  console.error("Service worker: Global error", event.error);
});

self.addEventListener("unhandledrejection", (event: PromiseRejectionEvent) => {
  console.error("Service worker: Unhandled promise rejection", event.reason);
});
