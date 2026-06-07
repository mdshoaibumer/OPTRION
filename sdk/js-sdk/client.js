/**
 * OPTRION JavaScript SDK
 * 
 * Usage:
 *   const optrion = require('@optrion/sdk');
 *   
 *   const client = new optrion.Client({
 *     endpoint: 'http://localhost:8080',
 *     apiKey: 'your-api-key',
 *     productId: 'your-product-id',
 *   });
 *   
 *   await client.register();
 *   await client.startMonitoring();
 */

const axios = require('axios');
const os = require('os');

/**
 * Configuration for OPTRION SDK
 */
class Config {
  constructor(options = {}) {
    this.endpoint = options.endpoint || 'http://localhost:8080';
    this.apiKey = options.apiKey || '';
    this.tenantId = options.tenantId || '';
    this.productId = options.productId || '';
    this.environmentId = options.environmentId || '';
    this.metricsInterval = options.metricsInterval || 30000; // 30 seconds
    this.healthCheckPath = options.healthCheckPath || '/health';
    this.collectors = options.collectors || ['runtime', 'memory', 'cpu'];
    this.autoDiscovery = options.autoDiscovery !== false;
    this.logger = options.logger || console;
  }
}

/**
 * OPTRION SDK Client for JavaScript/Node.js
 */
class Client {
  constructor(config) {
    if (!(config instanceof Config)) {
      config = new Config(config);
    }

    this.config = config;
    this.httpClient = axios.create({
      baseURL: config.endpoint,
      headers: {
        'Authorization': `Bearer ${config.apiKey}`,
        'Content-Type': 'application/json',
        'User-Agent': 'optrion-js-sdk/1.0.0',
      },
      timeout: 10000,
    });

    this.metricsInterval = null;
    this.isRunning = false;
    this.customCollectors = new Map();
  }

  /**
   * Register application with OPTRION platform
   */
  async register() {
    this.config.logger.log('Registering with OPTRION', {
      endpoint: this.config.endpoint,
      productId: this.config.productId,
    });

    try {
      await this.httpClient.post('/api/v1/register', {
        tenant_id: this.config.tenantId,
        product_id: this.config.productId,
        environment_id: this.config.environmentId,
        hostname: os.hostname(),
        node_version: process.version,
        platform: process.platform,
        arch: process.arch,
      });

      this.config.logger.log('Successfully registered with OPTRION');
      return true;
    } catch (error) {
      this.config.logger.error('Registration failed', error.message || error);
      throw error;
    }
  }

  /**
   * Start collecting and sending metrics
   */
  async startMonitoring() {
    if (this.isRunning) {
      this.config.logger.warn('Monitoring already running');
      return;
    }

    this.config.logger.log('Starting metrics collection', {
      interval: this.config.metricsInterval,
    });

    this.isRunning = true;

    // Collect metrics immediately
    await this.collectAndSendMetrics();

    // Then start periodic collection
    this.metricsInterval = setInterval(
      () => this.collectAndSendMetrics(),
      this.config.metricsInterval
    );

    this.config.logger.log('Metrics collection started');
  }

  /**
   * Stop collecting and sending metrics
   */
  stopMonitoring() {
    if (!this.isRunning) {
      return;
    }

    if (this.metricsInterval) {
      clearInterval(this.metricsInterval);
      this.metricsInterval = null;
    }

    this.isRunning = false;
    this.config.logger.log('Metrics collection stopped');
  }

  /**
   * Collect metrics and send to OPTRION
   */
  async collectAndSendMetrics() {
    try {
      const metrics = this.collectMetrics();

      // Collect custom metrics
      for (const [name, collector] of this.customCollectors) {
        try {
          const customMetrics = await collector.collect();
          for (const [key, value] of Object.entries(customMetrics)) {
            metrics[`${name}.${key}`] = value;
          }
        } catch (err) {
          this.config.logger.warn(`Custom collector '${name}' failed`, err.message || err);
        }
      }

      await this.httpClient.post('/api/v1/metrics', {
        tenant_id: this.config.tenantId,
        product_id: this.config.productId,
        environment_id: this.config.environmentId,
        timestamp: new Date().toISOString(),
        metrics,
      });

      this.config.logger.debug && this.config.logger.debug('Metrics collected and sent', {
        metric_count: Object.keys(metrics).length,
      });

      return metrics;
    } catch (error) {
      this.config.logger.error('Failed to send metrics', error.message || error);
    }
  }

  /**
   * Collect runtime metrics
   */
  collectMetrics() {
    const memUsage = process.memoryUsage();
    const cpuUsage = process.cpuUsage();

    return {
      timestamp: new Date().toISOString(),
      process: {
        pid: process.pid,
        uptime: process.uptime(),
        memory: {
          rss: memUsage.rss,
          heapTotal: memUsage.heapTotal,
          heapUsed: memUsage.heapUsed,
          external: memUsage.external,
          arrayBuffers: memUsage.arrayBuffers || 0,
        },
        cpu: {
          user: cpuUsage.user,
          system: cpuUsage.system,
        },
      },
      system: {
        arch: process.arch,
        platform: process.platform,
        nodeVersion: process.version,
        cpuCount: os.cpus().length,
        totalMemory: os.totalmem(),
        freeMemory: os.freemem(),
        loadAverage: os.loadavg(),
      },
    };
  }

  /**
   * Get current health status
   */
  async getHealth() {
    const metrics = this.collectMetrics();

    return {
      status: 'healthy',
      timestamp: new Date().toISOString(),
      metrics,
    };
  }

  /**
   * Register a custom metric collector
   * @param {string} name - Unique name for the collector
   * @param {{ collect: () => Promise<object> | object }} collector - Object with a collect() method
   */
  registerMetricCollector(name, collector) {
    if (!name || typeof name !== 'string') {
      throw new Error('Collector name must be a non-empty string');
    }
    if (!collector || typeof collector.collect !== 'function') {
      throw new Error('Collector must have a collect() method');
    }

    this.customCollectors.set(name, collector);
    this.config.logger.log('Metric collector registered', { name });
  }

  /**
   * Validate connection to OPTRION server
   */
  async validate() {
    try {
      const response = await this.httpClient.get('/healthz');
      return response.status === 200;
    } catch (error) {
      this.config.logger.error('Connection validation failed', error.message || error);
      return false;
    }
  }
}

/**
 * Exported SDK interface
 */
module.exports = {
  Client,
  Config,

  /**
   * Create a new OPTRION SDK client
   */
  create(options) {
    return new Client(options);
  },

  /**
   * Create configuration object
   */
  config(options) {
    return new Config(options);
  },

  /**
   * Version
   */
  version: '1.0.0',
};

// Export for CommonJS
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    Client,
    Config,
    create(options) {
      return new Client(options);
    },
    config(options) {
      return new Config(options);
    },
    version: '1.0.0',
  };
}
