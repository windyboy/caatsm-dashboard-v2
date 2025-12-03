/**
 * Health status utility functions
 */

export function getStatusColor(status: string): string {
  switch (status) {
    case "ok":
    case "healthy":
      return "text-green-800 bg-green-50 border-green-200";
    case "error":
    case "unhealthy":
      return "text-red-600 bg-red-50 border-red-200";
    case "not_configured":
    case "degraded":
      return "text-yellow-800 bg-yellow-50 border-yellow-200";
    default:
      return "text-gray-600 bg-gray-50 border-gray-200";
  }
}

export function getStatusIcon(status: string): string {
  switch (status) {
    case "ok":
    case "healthy":
      return "✓";
    case "error":
    case "unhealthy":
      return "✗";
    case "not_configured":
    case "degraded":
      return "○";
    default:
      return "?";
  }
}

export function getStatusLabel(status: string): string {
  switch (status) {
    case "ok":
    case "healthy":
      return "Healthy";
    case "error":
    case "unhealthy":
      return "Error";
    case "not_configured":
      return "Not Configured";
    case "degraded":
      return "Degraded";
    default:
      return "Unknown";
  }
}

export function formatUptime(uptime?: string): string {
  if (!uptime) return "";
  const seconds = parseInt(uptime, 10);
  if (isNaN(seconds)) return "";
  
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`;
  } else if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else {
    return `${minutes}m`;
  }
}

