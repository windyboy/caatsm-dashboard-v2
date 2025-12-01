/**
 * Hashes a string to a short identifier for logging purposes.
 * This prevents PII exposure in logs while still allowing correlation.
 *
 * @param str - The string to hash
 * @returns A short hash string (base36)
 */
export function hashString(str: string): string {
  if (!str || typeof str !== "string") {
    return "";
  }

  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }

  return hash.toString(36);
}

