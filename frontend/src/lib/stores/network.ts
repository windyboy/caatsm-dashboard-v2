// Network status store for offline detection

import { writable } from "svelte/store";

function createNetworkStore() {
  const { subscribe, set } = writable(typeof navigator !== "undefined" ? navigator.onLine : true);

  if (typeof window !== "undefined") {
    window.addEventListener("online", () => set(true));
    window.addEventListener("offline", () => set(false));
  }

  return { subscribe };
}

export const isOnline = createNetworkStore();

