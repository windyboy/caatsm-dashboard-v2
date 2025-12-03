/**
 * Browser environment utilities for SSR-safe code
 */

export const isBrowser = typeof window !== "undefined";

/**
 * Execute function only in browser environment
 */
export function browserOnly<T>(fn: () => T, fallback?: T): T | undefined {
  return isBrowser ? fn() : fallback;
}

/**
 * Safely access window object
 */
export function getWindow(): Window | undefined {
  return isBrowser ? window : undefined;
}

/**
 * Safely access document object
 */
export function getDocument(): Document | undefined {
  return isBrowser ? document : undefined;
}

