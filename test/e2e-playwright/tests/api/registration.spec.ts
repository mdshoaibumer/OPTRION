import { test, expect } from '../fixtures';

test.describe('Registration API', () => {
  test('POST /api/v1/register creates tenant successfully', async ({ apiContext }) => {
    const slug = `reg-test-${Date.now()}`;
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: {
          name: `Registration Test ${slug}`,
          slug: slug,
          plan: 'starter',
        },
        product: {
          name: 'Test Product',
          slug: `product-${slug}`,
          description: 'Playwright E2E test product',
          version: '1.0.0',
        },
        environment: {
          name: 'Development',
          tier: 'development',
        },
        components: [
          {
            name: 'Test Service',
            kind: 'api',
            description: 'A test API service',
            endpoint: 'http://localhost:4000',
            port: 4000,
          },
        ],
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body).toHaveProperty('tenant_id');
    expect(body).toHaveProperty('product_id');
    expect(body).toHaveProperty('environment_id');
    expect(body).toHaveProperty('component_ids');
    expect(body).toHaveProperty('api_key');
    expect(body).toHaveProperty('message');

    // Validate UUIDs format
    expect(body.tenant_id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
    );
    expect(body.api_key).toBeTruthy();
    expect(body.component_ids).toHaveLength(1);
  });

  test('POST /api/v1/register with multiple components', async ({ apiContext }) => {
    const slug = `multi-comp-${Date.now()}`;
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: {
          name: `Multi Component Test ${slug}`,
          slug: slug,
          plan: 'professional',
        },
        product: {
          name: 'Multi Service App',
          slug: `multi-${slug}`,
          description: 'App with multiple components',
          version: '2.0.0',
        },
        environment: {
          name: 'Staging',
          tier: 'staging',
        },
        components: [
          {
            name: 'API Gateway',
            kind: 'api',
            description: 'Main API gateway',
            endpoint: 'http://localhost:8000',
            port: 8000,
          },
          {
            name: 'PostgreSQL',
            kind: 'database',
            description: 'Primary database',
            endpoint: 'postgresql://localhost:5432/app',
            port: 5432,
          },
          {
            name: 'Redis Cache',
            kind: 'cache',
            description: 'Cache layer',
            endpoint: 'redis://localhost:6379',
            port: 6379,
          },
        ],
      },
    });

    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body.component_ids).toHaveLength(3);
  });

  test('POST /api/v1/register with duplicate slug returns 409', async ({ apiContext }) => {
    const slug = `dup-test-${Date.now()}`;
    const payload = {
      tenant: { name: `Dup Test ${slug}`, slug: slug, plan: 'starter' },
      product: { name: 'Dup Product', slug: `dup-prod-${slug}`, description: 'Dup', version: '1.0.0' },
      environment: { name: 'Dev', tier: 'development' },
      components: [{ name: 'Svc', kind: 'api', description: 'Svc', endpoint: 'http://localhost:9000', port: 9000 }],
    };

    // First registration should succeed
    const first = await apiContext.post('/api/v1/register', { data: payload });
    expect(first.status()).toBe(201);

    // Second registration with same slug should conflict
    const second = await apiContext.post('/api/v1/register', { data: payload });
    expect(second.status()).toBe(409);
  });

  test('POST /api/v1/register with invalid data returns 400', async ({ apiContext }) => {
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: { name: '', slug: '', plan: 'invalid-plan' },
        product: { name: '', slug: '' },
        environment: { name: '', tier: '' },
        components: [],
      },
    });

    expect(response.status()).toBe(400);
  });

  test('POST /api/v1/register with empty body returns 400', async ({ apiContext }) => {
    const response = await apiContext.post('/api/v1/register', {
      data: {},
    });

    expect(response.status()).toBe(400);
  });
});
