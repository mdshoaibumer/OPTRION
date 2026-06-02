import { defineConfig, devices } from '@playwright/test';
import * as path from 'path';
import * as fs from 'fs';

// Load .env file
const envPath = path.resolve(__dirname, '.env');
if (fs.existsSync(envPath)) {
  const envContent = fs.readFileSync(envPath, 'utf-8');
  for (const line of envContent.split('\n')) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) continue;
    const [key, ...valueParts] = trimmed.split('=');
    const value = valueParts.join('=');
    if (key && !process.env[key]) {
      process.env[key] = value;
    }
  }
}

/**
 * Optrion E2E Playwright Configuration
 * Supports both API testing and future browser-based UI testing.
 *
 * Run modes:
 *   npm test              → headless (all projects)
 *   npm run test:headed   → headed (browser visible)
 *   npm run test:api      → API-only tests (no browser)
 *   npm run test:ui       → Browser-based UI tests
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html', { open: 'never' }],
    ['list'],
  ],
  use: {
    baseURL: process.env.BASE_URL || 'http://localhost:8080',
    trace: 'on-first-retry',
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
    },
  },

  projects: [
    // API testing project - no browser needed
    {
      name: 'api',
      testDir: './tests/api',
      use: {
        baseURL: process.env.BASE_URL || 'http://localhost:8080',
      },
    },

    // Browser-based UI testing projects (for future dashboard/UI)
    {
      name: 'chromium',
      testDir: './tests/ui',
      use: {
        ...devices['Desktop Chrome'],
        baseURL: process.env.UI_BASE_URL || 'http://localhost:3000',
      },
    },
    {
      name: 'firefox',
      testDir: './tests/ui',
      use: {
        ...devices['Desktop Firefox'],
        baseURL: process.env.UI_BASE_URL || 'http://localhost:3000',
      },
    },
    {
      name: 'webkit',
      testDir: './tests/ui',
      use: {
        ...devices['Desktop Safari'],
        baseURL: process.env.UI_BASE_URL || 'http://localhost:3000',
      },
    },
  ],

  /* Start the Optrion API server before running tests (set SERVER_START=true to auto-start) */
  ...(process.env.SERVER_START === 'true'
    ? {
        webServer: {
          command: process.env.SERVER_CMD || 'go run ./testserver',
          url: 'http://localhost:8080/healthz',
          reuseExistingServer: !process.env.CI,
          timeout: 120000,
          stdout: 'pipe' as const,
          stderr: 'pipe' as const,
        },
      }
    : {}),
});
