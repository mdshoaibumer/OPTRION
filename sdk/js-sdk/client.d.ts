/**
 * OPTRION JavaScript SDK TypeScript Definitions
 */

export interface ConfigOptions {
  endpoint?: string;
  apiKey?: string;
  tenantId?: string;
  productId?: string;
  environmentId?: string;
  metricsInterval?: number;
  healthCheckPath?: string;
  collectors?: string[];
  autoDiscovery?: boolean;
  logger?: Logger;
}

export interface Logger {
  log(...args: any[]): void;
  error(...args: any[]): void;
  warn(...args: any[]): void;
  debug(...args: any[]): void;
}

export interface Metrics {
  timestamp: string;
  process: {
    pid: number;
    uptime: number;
    memory: {
      rss: number;
      heapTotal: number;
      heapUsed: number;
      external: number;
      arrayBuffers: number;
    };
    cpu: {
      user: number;
      system: number;
    };
  };
  system: {
    arch: string;
    platform: string;
    nodeVersion: string;
    cpuCount: number;
    totalMemory: number;
    freeMemory: number;
  };
}

export interface HealthStatus {
  status: string;
  timestamp: string;
  metrics: Metrics;
}

export class Config {
  endpoint: string;
  apiKey: string;
  tenantId: string;
  productId: string;
  environmentId: string;
  metricsInterval: number;
  healthCheckPath: string;
  collectors: string[];
  autoDiscovery: boolean;
  logger: Logger;

  constructor(options?: ConfigOptions);
}

export class Client {
  constructor(config: Config | ConfigOptions);

  /**
   * Register application with OPTRION platform
   */
  register(): Promise<boolean>;

  /**
   * Start collecting and sending metrics
   */
  startMonitoring(): Promise<void>;

  /**
   * Stop collecting and sending metrics
   */
  stopMonitoring(): void;

  /**
   * Collect metrics and send to OPTRION
   */
  collectAndSendMetrics(): Promise<Metrics>;

  /**
   * Collect runtime metrics
   */
  collectMetrics(): Metrics;

  /**
   * Get current health status
   */
  getHealth(): Promise<HealthStatus>;

  /**
   * Register a custom metric collector
   */
  registerMetricCollector(name: string, collector: MetricCollector): void;

  /**
   * Validate connection to OPTRION server
   */
  validate(): Promise<boolean>;
}

export interface MetricCollector {
  collect(): Promise<Record<string, any>>;
  name(): string;
}

export interface SDK {
  Client: typeof Client;
  Config: typeof Config;
  create(options: ConfigOptions): Client;
  config(options?: ConfigOptions): Config;
  version: string;
}

declare const sdk: SDK;
export default sdk;
