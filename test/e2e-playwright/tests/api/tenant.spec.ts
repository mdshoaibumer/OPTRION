import { test, expect } from '../fixtures';

test.describe('Tenant API (Authenticated)', () => {
  test('GET /api/v1/tenants returns tenant info', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/tenants');
    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    expect(body).toBeDefined();
  });

  test('GET /api/v1/products returns products for tenant', async ({ authedContext, registeredTenant }) => {
    const response = await authedContext.get(`/api/v1/products?tenant_id=${registeredTenant.tenantId}`);
    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    expect(body).toHaveProperty('data');
  });

  test('GET /api/v1/environments returns environments', async ({ authedContext, registeredTenant }) => {
    const response = await authedContext.get(`/api/v1/environments?product_id=${registeredTenant.productId}`);
    expect(response.ok()).toBeTruthy();
  });

  test('GET /api/v1/components returns components', async ({ authedContext, registeredTenant }) => {
    const response = await authedContext.get(`/api/v1/components?environment_id=${registeredTenant.environmentId}`);
    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    expect(body).toHaveProperty('data');
  });
});
