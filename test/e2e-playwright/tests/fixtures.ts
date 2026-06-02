import { test as base, APIRequestContext } from '@playwright/test';

/**
 * Custom fixtures for Optrion API testing.
 * Provides authenticated and unauthenticated API contexts.
 */
export type OptrionFixtures = {
  apiContext: APIRequestContext;
  authedContext: APIRequestContext;
  registeredTenant: {
    tenantId: string;
    productId: string;
    environmentId: string;
    componentIds: string[];
    apiKey: string;
  };
};

export const test = base.extend<OptrionFixtures>({
  // Unauthenticated API context
  apiContext: async ({ playwright }, use) => {
    const context = await playwright.request.newContext({
      baseURL: process.env.BASE_URL || 'http://localhost:8080',
      extraHTTPHeaders: {
        'Content-Type': 'application/json',
      },
    });
    await use(context);
    await context.dispose();
  },

  // Registers a fresh tenant and provides authenticated context
  registeredTenant: async ({ apiContext }, use) => {
    const slug = `e2e-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
    const response = await apiContext.post('/api/v1/register', {
      data: {
        tenant: {
          name: `E2E Test Tenant ${slug}`,
          slug: slug,
          plan: 'starter',
        },
        product: {
          name: 'E2E Test Product',
          slug: `product-${slug}`,
          description: 'Product created by Playwright E2E tests',
          version: '1.0.0',
        },
        environment: {
          name: 'Testing',
          tier: 'development',
        },
        components: [
          {
            name: 'Test API',
            kind: 'api',
            description: 'Test API component',
            endpoint: 'http://localhost:3000',
            port: 3000,
          },
          {
            name: 'Test Database',
            kind: 'database',
            description: 'Test database component',
            endpoint: 'postgresql://localhost:5432/testdb',
            port: 5432,
          },
        ],
      },
    });

    const body = await response.json();
    await use({
      tenantId: body.tenant_id,
      productId: body.product_id,
      environmentId: body.environment_id,
      componentIds: body.component_ids,
      apiKey: body.api_key,
    });
  },

  // Authenticated API context (uses a registered tenant's API key)
  authedContext: async ({ playwright, registeredTenant }, use) => {
    const context = await playwright.request.newContext({
      baseURL: process.env.BASE_URL || 'http://localhost:8080',
      extraHTTPHeaders: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${registeredTenant.apiKey}`,
      },
    });
    await use(context);
    await context.dispose();
  },
});

export { expect } from '@playwright/test';
