// Svelte store for managing time range selection

import { writable } from "svelte/store";

export type TimeRangePreset = "1h" | "24h" | "7d" | "30d" | "90d" | "custom";

export interface TimeRange {
  start: Date;
  end: Date;
  preset: TimeRangePreset;
}

function getPresetRange(preset: TimeRangePreset): { start: Date; end: Date } {
  const end = new Date();
  const start = new Date();

  switch (preset) {
    case "1h":
      start.setHours(start.getHours() - 1);
      break;
    case "24h":
      start.setHours(start.getHours() - 24);
      break;
    case "7d":
      start.setDate(start.getDate() - 7);
      break;
    case "30d":
      start.setDate(start.getDate() - 30);
      break;
    case "90d":
      start.setDate(start.getDate() - 90);
      break;
    case "custom":
      // Custom range - use provided dates
      return { start, end };
  }

  return { start, end };
}

const defaultPreset: TimeRangePreset = "24h";
const defaultRange = getPresetRange(defaultPreset);

const initialState: TimeRange = {
  start: defaultRange.start,
  end: defaultRange.end,
  preset: defaultPreset,
};

function createTimeRangeStore() {
  const { subscribe, set, update } = writable<TimeRange>(initialState);

  return {
    subscribe,
    /**
     * Sets the time range using a preset.
     */
    setPreset: (preset: TimeRangePreset) => {
      const range = getPresetRange(preset);
      set({
        start: range.start,
        end: range.end,
        preset,
      });
    },
    /**
     * Sets a custom time range.
     */
    setCustomRange: (start: Date, end: Date) => {
      set({
        start,
        end,
        preset: "custom",
      });
    },
    /**
     * Updates the time range.
     */
    setRange: (range: TimeRange) => {
      set(range);
    },
    /**
     * Resets to default (24h).
     */
    reset: () => {
      set(initialState);
    },
    /**
     * Gets ISO string format for API calls.
     */
    getISOStrings: (range: TimeRange): { start_time: string; end_time: string } => {
      return {
        start_time: range.start.toISOString(),
        end_time: range.end.toISOString(),
      };
    },
  };
}

export const timeRange = createTimeRangeStore();

