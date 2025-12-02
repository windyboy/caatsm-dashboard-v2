// Network status store for offline detection

import { writable } from "svelte/store";

function createNetworkStore() {
  const { subscribe, set } = writable(typeof navigator !== "undefined" ? navigator.onLine : true);

  let onlineHandler: (() => void) | null = null;
  let offlineHandler: (() => void) | null = null;

  if (typeof window !== "undefined") {
    onlineHandler = () => set(true);
    offlineHandler = () => set(false);
    window.addEventListener("online", onlineHandler);
    window.addEventListener("offline", offlineHandler);
  }

  return {
    subscribe,
    cleanup: () => {
      if (typeof window !== "undefined") {
        if (onlineHandler) {
          window.removeEventListener("online", onlineHandler);
          onlineHandler = null;
        }
        if (offlineHandler) {
          window.removeEventListener("offline", offlineHandler);
          offlineHandler = null;
        }
      }
    },
  };
}

export const isOnline = createNetworkStore();

