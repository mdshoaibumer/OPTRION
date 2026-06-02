import { test, expect } from "@playwright/test";

test.describe("Health Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/health");
    await page.waitForSelector("h1");
  });

  test("renders page title", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("Health Scores");
  });

  test("displays component health cards", async ({ page }) => {
    await expect(page.getByText("Backend API")).toBeVisible();
    await expect(page.getByText("PostgreSQL Primary")).toBeVisible();
    await expect(page.getByText("Redis Cache")).toBeVisible();
    await expect(page.getByText("Redis Sessions")).toBeVisible();
    await expect(page.getByText("NGINX Proxy")).toBeVisible();
  });

  test("displays health metrics for components", async ({ page }) => {
    await expect(page.getByText("Latency P99")).toBeVisible();
    await expect(page.getByText("Error Rate")).toBeVisible();
    await expect(page.getByText("Connections")).toBeVisible();
  });

  test("shows health scores", async ({ page }) => {
    // Shows metric values from the mock data
    await expect(page.getByText("245ms")).toBeVisible();
    await expect(page.getByText("24/25")).toBeVisible();
  });
});
