// Client-side hooks for SvelteKit
import { dev } from "$app/environment";
import { onCLS, onINP, onFCP, onLCP, onTTFB, type Metric } from 'web-vitals';

// Performance monitoring
function reportWebVitals(metric: Metric) {
  // Log performance metrics to console
  console.log(`Web Vitals - ${metric.name}:`, metric);

  // In production, you could send to analytics service
  if (!dev) {
    // Example: send to backend API
    // fetch('/api/performance', {
    //   method: 'POST',
    //   body: JSON.stringify(metric),
    //   headers: { 'Content-Type': 'application/json' }
    // }).catch(err => console.warn('Failed to report performance metric:', err));
  }
}

// Monitor core web vitals
if (typeof window !== 'undefined') {
  onCLS(reportWebVitals);
  onINP(reportWebVitals);
  onFCP(reportWebVitals);
  onLCP(reportWebVitals);
  onTTFB(reportWebVitals);
}

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
