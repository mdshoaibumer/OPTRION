import { test, expect } from '../fixtures';

test.describe('AI & Recommendations API (Authenticated)', () => {
  test('GET /api/v1/analysis without incident_id returns 400', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/analysis');
    // Requires incident_id query param
    expect(response.status()).toBe(400);

    const body = await response.json();
    expect(body).toHaveProperty('error');
  });

  test('GET /api/v1/recommendations without incident_id returns 400', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/recommendations');
    // Requires incident_id query param
    expect(response.status()).toBe(400);

    const body = await response.json();
    expect(body).toHaveProperty('error');
  });

  test('GET /api/v1/analysis with invalid incident_id returns 400', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/analysis?incident_id=not-a-uuid');
    expect(response.status()).toBe(400);
  });

  test('GET /api/v1/recommendations with invalid incident_id returns 400', async ({ authedContext }) => {
    const response = await authedContext.get('/api/v1/recommendations?incident_id=not-a-uuid');
    expect(response.status()).toBe(400);
  });
});

test.describe('AI & Recommendations API - Unauthenticated', () => {
  test('GET /api/v1/analysis without auth returns 401', async ({ apiContext }) => {
    const response = await apiContext.get('/api/v1/analysis');
    expect(response.status()).toBe(401);
  });

  test('GET /api/v1/recommendations without auth returns 401', async ({ apiContext }) => {
    const response = await apiContext.get('/api/v1/recommendations');
    expect(response.status()).toBe(401);
  });
});
