// Network status store for offline detection

import { writable } from "svelte/store";
import { isBrowser, getWindow } from "../utils/browser";

function createNetworkStore() {
  const { subscribe, set } = writable(
    isBrowser && typeof navigator !== "undefined" ? navigator.onLine : true
  );

  let onlineHandler: (() => void) | null = null;
  let offlineHandler: (() => void) | null = null;

  const win = getWindow();
  if (win) {
    onlineHandler = () => set(true);
    offlineHandler = () => set(false);
    win.addEventListener("online", onlineHandler);
    win.addEventListener("offline", offlineHandler);
  }

  return {
    subscribe,
    cleanup: () => {
      const win = getWindow();
      if (win) {
        if (onlineHandler) {
          win.removeEventListener("online", onlineHandler);
          onlineHandler = null;
        }
        if (offlineHandler) {
          win.removeEventListener("offline", offlineHandler);
          offlineHandler = null;
        }
      }
    },
  };
}

export const isOnline = createNetworkStore();

