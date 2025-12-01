// Svelte store for managing message list

import { writable } from "svelte/store";
import type { Telegram } from "../utils/types";
import { UI_CONFIG } from "../constants";

function createMessagesStore() {
  const { subscribe, update } = writable<Telegram[]>([]);

  return {
    subscribe,
    add: (message: Telegram) => {
      update((messages: Telegram[]) => {
        // Add new message at the beginning
        const newMessages = [message, ...messages];
        // Keep only the last MAX_MESSAGES
        return newMessages.slice(0, UI_CONFIG.MAX_MESSAGES);
      });
    },
    addMultiple: (newMessages: Telegram[]) => {
      update((messages: Telegram[]) => {
        // Add new messages at the beginning, preserving order
        const combined = [...newMessages, ...messages];
        // Keep only the last MAX_MESSAGES
        return combined.slice(0, UI_CONFIG.MAX_MESSAGES);
      });
    },
    clear: () => {
      update(() => []);
    },
  };
}

export const messages = createMessagesStore();
