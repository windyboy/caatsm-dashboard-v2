/**
 * Error tracking utility for frontend error monitoring.
 * Can be extended to integrate with services like Sentry in production.
 */

export interface ErrorContext {
  component?: string;
  action?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Logs an error with optional context.
 * In development, logs to console. In production, can send to error tracking service.
 */
export function trackError(error: Error, context?: ErrorContext): void {
  if (import.meta.env.DEV) {
    console.error("Error:", error.message, context);
    if (error.stack) {
      console.error("Stack:", error.stack);
    }
  } else {
    // Production: Send to error tracking service (e.g., Sentry)
    // Example: Sentry.captureException(error, { extra: context });
  }
}

/**
 * Tracks a non-error event (e.g., performance issue, warning).
 */
export function trackEvent(name: string, data?: Record<string, unknown>): void {
  if (import.meta.env.DEV) {
    console.log(`Event: ${name}`, data);
  } else {
    // Production: Send to analytics service
  }
}

