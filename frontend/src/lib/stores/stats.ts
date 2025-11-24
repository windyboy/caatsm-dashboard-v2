// Svelte store for managing statistics

import { writable } from "svelte/store";

export interface StatsState {
  total: number;
  byPriority: Record<number, number>;
  byType: Record<string, number>;
}

const initialState: StatsState = {
  total: 0,
  byPriority: {},
  byType: {},
};

function createStatsStore() {
  const { subscribe, set, update } = writable<StatsState>(initialState);

  return {
    subscribe,
    setTotal: (total: number) => {
      update((state) => ({ ...state, total }));
    },
    setByPriority: (byPriority: Record<number, number>) => {
      update((state) => ({ ...state, byPriority }));
    },
    setByType: (byType: Record<string, number>) => {
      update((state) => ({ ...state, byType }));
    },
    reset: () => {
      set(initialState);
    },
  };
}

export const stats = createStatsStore();

