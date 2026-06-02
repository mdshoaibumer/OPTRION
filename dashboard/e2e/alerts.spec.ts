import { test, expect } from "@playwright/test";

test.describe("Alerts Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/alerts");
    await page.waitForSelector("h1");
  });

  test("renders page title", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("Alert Feed");
  });

  test("displays alert cards", async ({ page }) => {
    await expect(
      page.getByText("CRITICAL: PostgreSQL pool exhausted")
    ).toBeVisible();
    await expect(page.getByText("MAJOR: API latency exceeds SLO")).toBeVisible();
    await expect(page.getByText("WARNING: Redis cache degraded")).toBeVisible();
  });

  test("shows severity filter buttons", async ({ page }) => {
    await expect(page.getByRole("button", { name: "critical" })).toBeVisible();
    await expect(page.getByRole("button", { name: "warning" })).toBeVisible();
  });
});
