import { test, expect } from '@playwright/test';

/**
 * Placeholder UI tests for future Optrion Dashboard.
 * These tests will be activated when a frontend/dashboard is implemented.
 *
 * Expected URL: http://localhost:3000 (or UI_BASE_URL env var)
 */

test.describe('Dashboard - Placeholder', () => {
  test.skip(true, 'UI not yet implemented - placeholder for future dashboard tests');

  test('dashboard loads and shows login page', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/Optrion/);
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
  });

  test('dashboard login with valid credentials', async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="email"]', 'admin@optrion.io');
    await page.fill('input[name="password"]', 'test-password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL(/dashboard/);
  });

  test('dashboard shows tenant overview', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page.locator('h1')).toContainText('Overview');
    await expect(page.locator('[data-testid="health-score"]')).toBeVisible();
  });

  test('dashboard shows component list', async ({ page }) => {
    await page.goto('/dashboard/components');
    await expect(page.locator('[data-testid="component-card"]')).toHaveCount(2);
  });

  test('dashboard shows incidents', async ({ page }) => {
    await page.goto('/dashboard/incidents');
    await expect(page.locator('[data-testid="incident-list"]')).toBeVisible();
  });

  test('dashboard shows alerts configuration', async ({ page }) => {
    await page.goto('/dashboard/alerts');
    await expect(page.locator('[data-testid="alert-rules"]')).toBeVisible();
  });
});
