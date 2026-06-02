import { test, expect } from '../fixtures';

test.describe('Security - Authentication', () => {
  test('authenticated endpoints reject missing auth header', async ({ apiContext }) => {
    const protectedEndpoints = [
      '/api/v1/alerts',
      '/api/v1/alert-rules',
      '/api/v1/escalation-policies',
      '/api/v1/analysis',
      '/api/v1/recommendations',
    ];

    for (const endpoint of protectedEndpoints) {
      const response = await apiContext.get(endpoint);
      expect(response.status(), `Expected 401 for ${endpoint}`).toBe(401);
    }
  });

  test('authenticated endpoints reject invalid API key', async ({ playwright }) => {
    const context = await playwright.request.newContext({
      baseURL: process.env.BASE_URL || 'http://localhost:8080',
      extraHTTPHeaders: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer invalid_key_12345',
      },
    });

    const response = await context.get('/api/v1/alerts');
    expect(response.status()).toBe(401);

    await context.dispose();
  });

  test('public endpoints are accessible without auth', async ({ apiContext }) => {
    const publicEndpoints = ['/healthz', '/readyz', '/api/v1/info'];

    for (const endpoint of publicEndpoints) {
      const response = await apiContext.get(endpoint);
      expect(response.ok(), `Expected 2xx for ${endpoint}, got ${response.status()}`).toBeTruthy();
    }
  });
});

test.describe('Security - Input Validation', () => {
  test('registration rejects XSS in tenant name', async ({ apiContext }) => {
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: {
          name: '<script>alert("xss")</script>',
          slug: `xss-test-${Date.now()}`,
          plan: 'starter',
        },
        product: { name: 'XSS Test', slug: `xss-prod-${Date.now()}`, description: 'Test', version: '1.0.0' },
        environment: { name: 'Dev', tier: 'development' },
        components: [{ name: 'Svc', kind: 'api', description: 'Svc', endpoint: 'http://localhost:9000', port: 9000 }],
      },
    });

    // Should either reject (400) or sanitize the input
    if (response.status() === 201) {
      const body = await response.json();
      // If accepted, the stored name should NOT contain raw script tags
      expect(body.message).not.toContain('<script>');
    } else {
      expect(response.status()).toBe(400);
    }
  });

  test('registration rejects SQL injection in slug', async ({ apiContext }) => {
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: {
          name: 'SQL Injection Test',
          slug: "'; DROP TABLE tenants; --",
          plan: 'starter',
        },
        product: { name: 'SQLi', slug: 'sqli-prod', description: 'Test', version: '1.0.0' },
        environment: { name: 'Dev', tier: 'development' },
        components: [{ name: 'Svc', kind: 'api', description: 'Svc', endpoint: 'http://localhost:9000', port: 9000 }],
      },
    });

    // Should reject invalid slug format
    expect(response.status()).toBe(400);
  });

  test('oversized payload is rejected', async ({ apiContext }) => {
    const largeString = 'x'.repeat(10 * 1024 * 1024); // 10MB
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: { name: largeString, slug: 'oversized', plan: 'starter' },
        product: { name: 'Big', slug: 'big', description: largeString, version: '1.0.0' },
        environment: { name: 'Dev', tier: 'development' },
        components: [],
      },
    });

    // Should reject with 400 or 413 (payload too large)
    expect([400, 413]).toContain(response.status());
  });
});
