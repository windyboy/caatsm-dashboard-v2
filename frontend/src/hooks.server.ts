// Server-side hooks for SvelteKit

import type { Handle } from "@sveltejs/kit";

/**
 * Content Security Policy configuration
 * In production, remove 'unsafe-inline' for script-src and style-src
 */
const getCSPDirectives = (isDev: boolean) => {
  const scriptSrc = isDev
    ? ["'self'", "'unsafe-inline'", "'unsafe-eval'"] // Required for HMR in dev
    : ["'self'"]; // Production: no unsafe-inline

  const styleSrc = isDev
    ? ["'self'", "'unsafe-inline'"] // Required for HMR in dev
    : ["'self'"]; // Production: no unsafe-inline

  return {
    "default-src": ["'self'"],
    "script-src": scriptSrc,
    "style-src": styleSrc,
    "connect-src": [
      "'self'",
      "ws://localhost:3002",
      "wss://*",
      "http://localhost:3002",
      "https://*",
    ],
    "img-src": ["'self'", "data:", "https:"],
    "font-src": ["'self'", "data:"],
    "object-src": ["'none'"],
    "base-uri": ["'self'"],
    "form-action": ["'self'"],
    "frame-ancestors": ["'none'"],
    "upgrade-insecure-requests": [],
  };
};

const formatCSP = (directives: Record<string, string[]>) => {
  return Object.entries(directives)
    .map(([key, values]) => {
      if (values.length === 0) {
        return key;
      }
      return `${key} ${values.join(" ")}`;
    })
    .join("; ");
};

export const handle: Handle = async ({ event, resolve }) => {
  const isDev = import.meta.env.DEV;
  const cspDirectives = getCSPDirectives(isDev);
  const cspHeader = formatCSP(cspDirectives);

  const response = await resolve(event);

  // Set CSP header
  response.headers.set("Content-Security-Policy", cspHeader);

  // Additional security headers
  response.headers.set("X-Content-Type-Options", "nosniff");
  response.headers.set("X-Frame-Options", "DENY");
  response.headers.set("X-XSS-Protection", "1; mode=block");
  response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");

  return response;
};
