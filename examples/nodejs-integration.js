/**
 * Example Node.js Integration with OPTRION
 * 
 * This example demonstrates how to integrate a Node.js/Express application
 * with the OPTRION platform for automatic health monitoring.
 * 
 * Usage:
 *   export OPTRION_ENDPOINT="http://localhost:8080"
 *   export OPTRION_API_KEY="optrion_key_xxxxx"
 *   export OPTRION_TENANT_ID="tenant-uuid"
 *   export OPTRION_PRODUCT_ID="product-uuid"
 *   export OPTRION_ENVIRONMENT_ID="env-uuid"
 *   node examples/nodejs-integration.js
 */

const express = require('express');
const optrion = require('@optrion/sdk');
const os = require('os');

// Initialize Express app
const app = express();
const port = process.env.PORT || 3000;

// Initialize OPTRION SDK
async function initializeOPTRION() {
  console.log('Initializing OPTRION...');

  // Create OPTRION SDK client
  const client = new optrion.Client({
    // OPTRION server endpoint
    endpoint: process.env.OPTRION_ENDPOINT || 'http://localhost:8080',

    // API key from registration (required)
    apiKey: process.env.OPTRION_API_KEY || '',

    // Application identifiers
    tenantId: process.env.OPTRION_TENANT_ID || '',
    productId: process.env.OPTRION_PRODUCT_ID || '',
    environmentId: process.env.OPTRION_ENVIRONMENT_ID || '',

    // Metrics collection interval
    metricsInterval: 30000, // 30 seconds

    // Health check endpoint
    healthCheckPath: '/health',

    // Collectors to enable
    collectors: ['runtime', 'memory', 'cpu', 'disk'],

    // Logger
    logger: console,
  });

  // Validate connection to OPTRION server
  const isConnected = await client.validate();
  if (!isConnected) {
    console.warn('Warning: Could not connect to OPTRION server');
  }

  // Register application with OPTRION
  try {
    const registered = await client.register();
    if (registered) {
      console.log('✓ Successfully registered with OPTRION');
    }
  } catch (error) {
    console.error('Registration failed:', error);
  }

  // Start metrics collection
  try {
    await client.startMonitoring();
    console.log('✓ Started OPTRION metrics collection');
  } catch (error) {
    console.error('Failed to start monitoring:', error);
  }

  // Setup graceful shutdown
  process.on('SIGTERM', () => {
    console.log('SIGTERM signal received: closing HTTP server');
    client.stopMonitoring();
    process.exit(0);
  });

  process.on('SIGINT', () => {
    console.log('SIGINT signal received: closing HTTP server');
    client.stopMonitoring();
    process.exit(0);
  });

  return client;
}

// Express middleware for request logging
app.use((req, res, next) => {
  console.log(`${req.method} ${req.path}`);
  next();
});

// Health check endpoint - used by OPTRION
app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

// Metrics endpoint
app.get('/metrics', (req, res) => {
  const memUsage = process.memoryUsage();
  const uptime = process.uptime();

  res.json({
    timestamp: new Date().toISOString(),
    uptime: uptime,
    memory: {
      rss: memUsage.rss,
      heapTotal: memUsage.heapTotal,
      heapUsed: memUsage.heapUsed,
      external: memUsage.external,
    },
    system: {
      platform: process.platform,
      arch: process.arch,
      cpuCount: os.cpus().length,
      totalMemory: os.totalmem(),
      freeMemory: os.freemem(),
    },
  });
});

// API endpoints
app.get('/api/users', (req, res) => {
  res.json([
    { id: 1, name: 'John Doe', email: 'john@example.com' },
    { id: 2, name: 'Jane Smith', email: 'jane@example.com' },
  ]);
});

app.get('/api/users/:id', (req, res) => {
  const userId = parseInt(req.params.id);
  res.json({ id: userId, name: 'User ' + userId, email: `user${userId}@example.com` });
});

// Error handling
app.use((err, req, res, next) => {
  console.error('Error:', err);
  res.status(500).json({ error: 'Internal Server Error' });
});

// Start server
async function start() {
  try {
    // Initialize OPTRION
    const optrionClient = await initializeOPTRION();

    // Start HTTP server
    app.listen(port, () => {
      console.log(`✓ Express server listening on port ${port}`);
      console.log(`✓ Health check: http://localhost:${port}/health`);
      console.log(`✓ Metrics: http://localhost:${port}/metrics`);
      console.log(`✓ API: http://localhost:${port}/api/users`);
    });
  } catch (error) {
    console.error('Failed to start server:', error);
    process.exit(1);
  }
}

// Start the application
start();

/**
 * Complete Setup Steps:
 * 
 * 1. Initialize OPTRION configuration:
 *    $ optrion-cli init
 * 
 * 2. Edit optrion.yaml with your application details
 * 
 * 3. Register with OPTRION server:
 *    $ optrion-cli register --config optrion.yaml --server http://localhost:8080
 * 
 * 4. Set environment variables with values from registration:
 *    $ export OPTRION_ENDPOINT="http://localhost:8080"
 *    $ export OPTRION_API_KEY="optrion_key_xxxxx"
 *    $ export OPTRION_TENANT_ID="tenant-uuid"
 *    $ export OPTRION_PRODUCT_ID="product-uuid"
 *    $ export OPTRION_ENVIRONMENT_ID="env-uuid"
 * 
 * 5. Install dependencies:
 *    $ npm install express @optrion/sdk
 * 
 * 6. Run the application:
 *    $ node examples/nodejs-integration.js
 * 
 * 7. Verify integration in another terminal:
 *    $ optrion-cli verify \
 *        --config optrion.yaml \
 *        --server http://localhost:8080 \
 *        --api-key optrion_key_xxxxx
 * 
 * Expected Output:
 *   ✓ Configuration file valid: OK
 *   ✓ Component Connectivity: OK
 *   ✓ Metrics Status: Flowing
 *   ✓ Integration verified successfully!
 */
