import type { Telegram } from "$lib/types";

/**
 * Checks if two telegrams are duplicates by comparing message_id first,
 * then falling back to content + time + type comparison.
 */
export function areTelegramsDuplicate(msg1: Telegram, msg2: Telegram): boolean {
  // Primary check: If both have message_id, compare by ID
  if (msg1.message_id && msg2.message_id) {
    return msg1.message_id === msg2.message_id;
  }

  // Secondary check: Compare by content+time+type
  const contentMatch = isEmptyOrEqual(msg1.content, msg2.content);
  const timeMatch = isEmptyOrEqual(msg1.time, msg2.time);
  const typeMatch = isEmptyOrEqual(msg1.type, msg2.type);

  return contentMatch && timeMatch && typeMatch;
}

/**
 * Checks if a telegram is a duplicate of any message in the array.
 */
export function isDuplicateMessage(telegram: Telegram, messages: Telegram[]): boolean {
  return messages.some((msg) => areTelegramsDuplicate(msg, telegram));
}

/**
 * Helper to check if two values are both empty or equal.
 */
function isEmptyOrEqual(val1: string | undefined | null, val2: string | undefined | null): boolean {
  const isEmpty1 = val1 === undefined || val1 === null || val1 === "";
  const isEmpty2 = val2 === undefined || val2 === null || val2 === "";

  if (isEmpty1 && isEmpty2) return true;
  if (isEmpty1 || isEmpty2) return false;

  return val1 === val2;
}
