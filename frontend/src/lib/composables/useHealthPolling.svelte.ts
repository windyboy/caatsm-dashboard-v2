/**
 * Composable for health status polling
 */

import { getHealthStatus, type HealthCheckResult } from "$lib/services/api";
import { createLogger } from "$lib/utils/logger";
import { isBrowser } from "$lib/utils/browser";

const logger = createLogger("useHealthPolling");

const POLL_INTERVAL_MS = 30000; // 30 seconds

export function useHealthPolling() {
  let healthData = $state<HealthCheckResult | null>(null);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let lastCheckTime = $state<Date | null>(null);
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  async function fetchHealthStatus() {
    if (loading) return;

    loading = true;
    error = null;

    try {
      const result = await getHealthStatus();
      healthData = result;
      lastCheckTime = new Date();
      logger.debug("Health status fetched", { status: result.status });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to fetch health status";
      error = errorMessage;
      logger.error("Failed to fetch health status", err);
    } finally {
      loading = false;
    }
  }

  function refresh() {
    fetchHealthStatus();
  }

  // Start polling using $effect
  $effect(() => {
    if (!isBrowser) {
      return;
    }

    // Initial fetch
    fetchHealthStatus();

    // Set up polling
    pollInterval = setInterval(() => {
      fetchHealthStatus();
    }, POLL_INTERVAL_MS);

    // Cleanup polling interval
    return () => {
      if (pollInterval) {
        clearInterval(pollInterval);
        pollInterval = null;
      }
    };
  });

  return {
    get healthData() {
      return healthData;
    },
    get loading() {
      return loading;
    },
    get error() {
      return error;
    },
    get lastCheckTime() {
      return lastCheckTime;
    },
    refresh,
  };
}

