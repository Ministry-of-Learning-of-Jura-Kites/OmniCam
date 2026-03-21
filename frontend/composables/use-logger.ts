import { getCurrentInstance } from "vue";

class Logger {
  constructor(private componentName: string = "Unknown") {}

  info(...args: unknown[]) {
    console.log(`[${this.componentName}]`, ...args);
  }

  error(...args: unknown[]) {
    console.error(`[${this.componentName}]`, ...args);
  }

  debug(...args: unknown[]) {
    console.warn(`[${this.componentName}]`, ...args);
  }
}

export function useLogger(name?: string) {
  if (name == null) {
    const instance = getCurrentInstance();
    name = instance?.type.__name || instance?.type.name || "Anonymous";
  }

  const logger = new Logger(name);

  return { logger };
}
