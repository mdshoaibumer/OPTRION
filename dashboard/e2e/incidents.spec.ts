import { test, expect } from "@playwright/test";

test.describe("Incidents Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/incidents");
    await page.waitForSelector("h1");
  });

  test("renders page title", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("Incident War Room");
    await expect(
      page.getByText("Manage active incidents")
    ).toBeVisible();
  });

  test("displays incident list", async ({ page }) => {
    await expect(
      page.getByText("PostgreSQL connection pool exhausted")
    ).toBeVisible();
    await expect(
      page.getByText("Redis cache hit ratio below threshold")
    ).toBeVisible();
    await expect(
      page.getByText("API response latency spike")
    ).toBeVisible();
  });

  test("can select an incident to view details", async ({ page }) => {
    await page.getByText("PostgreSQL connection pool exhausted").click();
    await expect(
      page.getByText("Connection pool utilization exceeded 95%")
    ).toBeVisible();
  });
});
