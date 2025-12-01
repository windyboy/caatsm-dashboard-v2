// Performance monitoring using Web Vitals

import { onCLS, onINP, onLCP, onFCP, onTTFB, type Metric } from "web-vitals";
import { createLogger } from "./logger";

const logger = createLogger("Performance");

/**
 * Reports a performance metric to the logger.
 * In production, this could also send metrics to an analytics service.
 */
function reportMetric(metric: Metric) {
  const { name, value, id, rating } = metric;

  logger.info("Web Vital", {
    metric: name,
    value: Math.round(value),
    id,
    rating,
  });

  // In production, you could send to analytics:
  // if (typeof window !== 'undefined' && window.gtag) {
  //   window.gtag('event', name, {
  //     value: Math.round(value),
  //     metric_id: id,
  //     metric_value: value,
  //     metric_delta: metric.delta,
  //   });
  // }
}

/**
 * Initializes Web Vitals performance monitoring.
 * Should be called once when the app loads.
 * Only runs in the browser (not during SSR).
 */
export function initPerformanceMonitoring() {
  if (typeof window === "undefined") {
    return;
  }

  try {
    onCLS(reportMetric);
    onINP(reportMetric);
    onLCP(reportMetric);
    onFCP(reportMetric);
    onTTFB(reportMetric);
  } catch (error) {
    logger.error("Failed to initialize performance monitoring", error);
  }
}

