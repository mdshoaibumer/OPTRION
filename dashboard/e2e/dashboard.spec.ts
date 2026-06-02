import { test, expect } from "@playwright/test";

test.describe("Dashboard - Home Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("h1");
  });

  test("renders the main dashboard with title and stats", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("Engineering Intelligence");
    await expect(page.getByText("Real-time health monitoring")).toBeVisible();
  });

  test("displays system health ring", async ({ page }) => {
    await expect(page.getByText("System Health")).toBeVisible();
  });

  test("displays incident list on dashboard", async ({ page }) => {
    await expect(
      page.getByText("PostgreSQL connection pool exhausted")
    ).toBeVisible();
  });

  test("displays alert feed on dashboard", async ({ page }) => {
    await expect(
      page.getByText("CRITICAL: PostgreSQL pool exhausted")
    ).toBeVisible();
  });

  test("displays stats cards", async ({ page }) => {
    await expect(page.getByText("Total Components")).toBeVisible();
    await expect(page.getByText("Avg Health Score")).toBeVisible();
  });
});
