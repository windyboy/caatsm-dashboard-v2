// Svelte store for managing message list

import { writable } from "svelte/store";
import type { Telegram } from "../utils/types";
import { UI_CONFIG } from "../constants";

/**
 * Checks if two messages are duplicates.
 * Uses message_id if available, otherwise compares multiple fields.
 */
function isDuplicateMessage(msg1: Telegram, msg2: Telegram): boolean {
  // Primary check: use message_id if both have valid message_ids
  if (msg1.message_id && msg1.message_id !== "0" && msg2.message_id && msg2.message_id !== "0") {
    return msg1.message_id === msg2.message_id;
  }

  // Fallback: compare multiple fields for messages without message_id
  return (
    msg1.time === msg2.time &&
    msg1.flight_number === msg2.flight_number &&
    msg1.source === msg2.source &&
    msg1.destination === msg2.destination &&
    msg1.type === msg2.type &&
    msg1.priority === msg2.priority &&
    msg1.content === msg2.content
  );
}

function createMessagesStore() {
  const { subscribe, update } = writable<Telegram[]>([]);

  return {
    subscribe,
    add: (message: Telegram) => {
      update((messages: Telegram[]) => {
        // Check if message already exists to prevent duplicates
        const isDuplicate = messages.some((existing) => isDuplicateMessage(message, existing));
        if (isDuplicate) {
          return messages; // Return unchanged if duplicate
        }

        // Add new message at the beginning
        const newMessages = [message, ...messages];
        // Keep only the last MAX_MESSAGES
        return newMessages.slice(0, UI_CONFIG.MAX_MESSAGES);
      });
    },
    addMultiple: (newMessages: Telegram[]) => {
      update((messages: Telegram[]) => {
        // Filter out duplicates from new messages
        const uniqueNewMessages = newMessages.filter(
          (newMsg) => !messages.some((existing) => isDuplicateMessage(newMsg, existing))
        );

        // Also remove duplicates within newMessages array itself
        const deduplicatedNewMessages: Telegram[] = [];
        for (const newMsg of uniqueNewMessages) {
          const isDuplicateInNew = deduplicatedNewMessages.some((existing) =>
            isDuplicateMessage(newMsg, existing)
          );
          if (!isDuplicateInNew) {
            deduplicatedNewMessages.push(newMsg);
          }
        }

        // Add new messages at the beginning, preserving order
        const combined = [...deduplicatedNewMessages, ...messages];
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
