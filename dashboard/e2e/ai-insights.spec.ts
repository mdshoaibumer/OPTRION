import { test, expect } from "@playwright/test";

test.describe("AI Insights Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/ai");
    await page.waitForSelector("h1");
  });

  test("renders page title", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("AI Insights");
  });

  test("displays AI insight cards", async ({ page }) => {
    await expect(
      page.getByText("PostgreSQL connection pool exhausted")
    ).toBeVisible();
    await expect(
      page.getByText("Redis cache hit ratio below threshold")
    ).toBeVisible();
  });

  test("shows root cause analysis", async ({ page }) => {
    await expect(
      page.getByText("Long-running transaction")
    ).toBeVisible();
  });

  test("shows confidence scores", async ({ page }) => {
    await expect(page.getByText("87%")).toBeVisible();
    await expect(page.getByText("72%")).toBeVisible();
  });

  test("displays recommendations", async ({ page }) => {
    await expect(
      page.getByText("Add composite index")
    ).toBeVisible();
    await expect(
      page.getByText("Increase maxmemory")
    ).toBeVisible();
  });
});
