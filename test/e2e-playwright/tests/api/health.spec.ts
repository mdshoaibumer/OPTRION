import { test, expect } from '../fixtures';

test.describe('Health Endpoints', () => {
  test('GET /healthz returns liveness status', async ({ apiContext }) => {
    const response = await apiContext.get('/healthz');
    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body).toHaveProperty('status');
    expect(body.status).toBe('alive');
  });

  test('GET /readyz returns readiness status', async ({ apiContext }) => {
    const response = await apiContext.get('/readyz');
    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body).toHaveProperty('status');
  });

  test('GET /api/v1/info returns API version info', async ({ apiContext }) => {
    const response = await apiContext.get('/api/v1/info');
    expect(response.status()).toBe(200);

    const body = await response.json();
    expect(body).toHaveProperty('version');
  });
});
