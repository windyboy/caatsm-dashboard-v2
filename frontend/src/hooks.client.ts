// Client-side hooks for SvelteKit
import { dev } from "$app/environment";

// Service worker management
if (typeof navigator !== "undefined" && "serviceWorker" in navigator) {
  if (dev) {
    // Unregister any existing service workers in development
    navigator.serviceWorker
      .getRegistrations()
      .then((registrations) => {
        for (const registration of registrations) {
          registration
            .unregister()
            .then((success) => {
              if (success) {
                console.debug("Service worker unregistered (development mode)");
              }
            })
            .catch((error) => {
              console.warn("Failed to unregister service worker", error);
            });
        }
      })
      .catch((error) => {
        console.warn("Failed to get service worker registrations", error);
      });
  } else {
    // In production, add error event listeners for debugging
    navigator.serviceWorker.addEventListener("error", (event) => {
      console.error("Service worker error:", event);
    });

    navigator.serviceWorker.addEventListener("messageerror", (event) => {
      console.error("Service worker message error:", event);
    });

    // Log service worker state changes
    navigator.serviceWorker.addEventListener("controllerchange", () => {
      console.debug("Service worker controller changed");
    });
  }
}
