import type { Handle } from "@sveltejs/kit";
import { init, addMessages } from "svelte-i18n";
// Messages are also added in app.ts, but we add them here too to ensure they're available
// before init() is called in the server hook for each request
import en from "./lib/i18n/locales/en.json";
import zh from "./lib/i18n/locales/zh.json";

export const handle: Handle = async ({ event, resolve }) => {
  // Ensure messages are added before initializing (messages from app.ts should already be added,
  // but this ensures they're available even if module loading order differs)
  // Register multiple locale keys to support browser standard formats
  addMessages("en", en);
  addMessages("en-US", en); // Compatible with browser
  addMessages("zh", zh);
  addMessages("zh-CN", zh); // Compatible with browser

  // Initialize i18n on server side
  // Extract locale from cookie, Accept-Language header, or use default
  let locale = event.cookies.get("locale");

  // If no valid cookie, check Accept-Language header
  if (!locale || !["en", "zh"].includes(locale)) {
    const acceptLanguage = event.request.headers.get("accept-language");
    if (acceptLanguage) {
      // Parse Accept-Language header (e.g., "en-US,en;q=0.9,zh;q=0.8")
      const preferredLocale = acceptLanguage.split(",")[0]?.split("-")[0]?.toLowerCase();
      if (preferredLocale && ["en", "zh"].includes(preferredLocale)) {
        locale = preferredLocale;
      }
    }
    // Fallback to 'en' if still no valid locale
    if (!locale || !["en", "zh"].includes(locale)) {
      locale = "en";
    }
  }

  // Store locale in cookie for client-side synchronization
  event.cookies.set("locale", locale, {
    path: "/",
    sameSite: "lax",
    maxAge: 60 * 60 * 24 * 365, // 1 year
  });

  // Initialize i18n with the determined locale
  // Messages are already added above, so init() can safely use them
  init({
    fallbackLocale: "en",
    initialLocale: locale,
  });

  const response = await resolve(event);

  // 添加安全头
  // Get API base URL from environment or use default
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || "http://localhost:3002";
  const isDev = import.meta.env.DEV;

  // Build connect-src directive
  // In development, allow localhost:3002 for backend API
  // In production, use the configured API URL
  const connectSrc = isDev
    ? "'self' ws: wss: http://localhost:3002 ws://localhost:3002"
    : `'self' ws: wss: ${apiBaseUrl} ${apiBaseUrl.replace("http://", "ws://").replace("https://", "wss://")}`;

  const securityHeaders = {
    // Content Security Policy
    // Note: 'unsafe-inline' and 'unsafe-eval' are required for SvelteKit in development
    // In production, consider using nonces for stricter CSP
    "Content-Security-Policy": [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'", // Required for SvelteKit HMR and hydration
      "style-src 'self' 'unsafe-inline'", // Required for Svelte scoped styles
      "img-src 'self' data: https:",
      "font-src 'self'",
      `connect-src ${connectSrc}`, // Allow connections to backend API
      "frame-ancestors 'none'",
      "base-uri 'self'",
      "form-action 'self'",
    ].join("; "),

    // 防止点击劫持
    "X-Frame-Options": "DENY",

    // 防止MIME类型混淆
    "X-Content-Type-Options": "nosniff",

    // XSS保护
    "X-XSS-Protection": "1; mode=block",

    // 引用策略
    "Referrer-Policy": "strict-origin-when-cross-origin",

    // 权限策略
    "Permissions-Policy": ["camera=()", "microphone=()", "geolocation=()", "payment=()"].join(", "),

    // HSTS (仅HTTPS)
    ...(event.url.protocol === "https:" && {
      "Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
    }),
  };

  // 设置安全头
  Object.entries(securityHeaders).forEach(([key, value]) => {
    if (value) {
      response.headers.set(key, value);
    }
  });

  return response;
};
