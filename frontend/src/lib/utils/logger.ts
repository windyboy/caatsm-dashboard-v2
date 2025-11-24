// Structured logger utility for frontend logging
// Provides consistent, filterable logging with levels and context

type LogLevel = "debug" | "info" | "warn" | "error";

interface LogContext {
  [key: string]: unknown;
}

function safeStringify(value: unknown): string {
  try {
    const seen = new WeakSet();
    return JSON.stringify(value, (_key, val) => {
      if (typeof val === "object" && val !== null) {
        if (seen.has(val)) {
          return "[Circular]";
        }
        seen.add(val);
      }
      if (typeof val === "function") {
        return "[Function]";
      }
      if (typeof val === "symbol") {
        return val.toString();
      }
      return val;
    });
  } catch (_err) {
    try {
      return String(value);
    } catch {
      return "[Non-serializable]";
    }
  }
}

class Logger {
  private prefix: string;
  private enabled: boolean;
  private minLevel: LogLevel;

  constructor(prefix: string, enabled = true, minLevel: LogLevel = "info") {
    this.prefix = prefix;
    this.enabled = enabled;
    this.minLevel = minLevel;
  }

  private shouldLog(level: LogLevel): boolean {
    if (!this.enabled) return false;

    const levels: LogLevel[] = ["debug", "info", "warn", "error"];
    return levels.indexOf(level) >= levels.indexOf(this.minLevel);
  }

  private formatMessage(level: LogLevel, message: string, context?: LogContext): string {
    const timestamp = new Date().toISOString();
    const contextStr = context ? ` ${safeStringify(context)}` : "";
    return `[${timestamp}] [${this.prefix}] [${level.toUpperCase()}] ${message}${contextStr}`;
  }

  debug(message: string, context?: LogContext): void {
    if (this.shouldLog("debug")) {
      console.debug(this.formatMessage("debug", message, context));
    }
  }

  info(message: string, context?: LogContext): void {
    if (this.shouldLog("info")) {
      console.info(this.formatMessage("info", message, context));
    }
  }

  warn(message: string, context?: LogContext): void {
    if (this.shouldLog("warn")) {
      console.warn(this.formatMessage("warn", message, context));
    }
  }

  error(message: string, error?: Error | unknown, context?: LogContext): void {
    if (this.shouldLog("error")) {
      const errorContext = {
        ...context,
        error: error instanceof Error ? {
          name: error.name,
          message: error.message,
          stack: error.stack,
        } : error,
      };
      console.error(this.formatMessage("error", message, errorContext));
    }
  }
}

// Create logger instances for different modules
export const createLogger = (prefix: string): Logger => {
  const isDev = import.meta.env.DEV;
  const minLevel = isDev ? "debug" : "info";
  return new Logger(prefix, true, minLevel);
};

// Export default logger for general use
export const logger = createLogger("App");

