import { test, expect } from '../fixtures';

test.describe('Alert API (Authenticated)', () => {
  test('GET /api/v1/alerts returns alerts list', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/alerts');
    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    expect(body).toBeDefined();
  });

  test('GET /api/v1/alert-rules returns alert rules', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/alert-rules');
    expect(response.ok()).toBeTruthy();
  });

  test('GET /api/v1/escalation-policies returns escalation policies', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/escalation-policies');
    expect(response.ok()).toBeTruthy();
  });
});

test.describe('Alert API - Unauthenticated', () => {
  test('GET /api/v1/alerts without auth returns 401', async ({ apiContext }) => {
    const response = await apiContext.get('/api/v1/alerts');
    expect(response.status()).toBe(401);
  });

  test('GET /api/v1/alert-rules without auth returns 401', async ({ apiContext }) => {
    const response = await apiContext.get('/api/v1/alert-rules');
    expect(response.status()).toBe(401);
  });
});
