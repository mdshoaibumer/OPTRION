import { test, expect } from "@playwright/test";

test.describe("Navigation", () => {
  test("sidebar renders all navigation items", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    const nav = page.locator("aside");
    await expect(nav).toBeVisible();
    await expect(nav.locator("text=OPTRION")).toBeVisible();
  });

  test("navigates to Topology page", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    await page.locator('a[href="/topology"]').click();
    await page.waitForURL("**/topology");
    await expect(page.locator("h1")).toContainText("Infrastructure Topology");
  });

  test("navigates to Incidents page", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    await page.locator('a[href="/incidents"]').click();
    await page.waitForURL("**/incidents");
    await expect(page.locator("h1")).toContainText("Incident War Room");
  });

  test("navigates to Health page", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    await page.locator('a[href="/health"]').click();
    await page.waitForURL("**/health");
    await expect(page.locator("h1")).toContainText("Health Scores");
  });

  test("navigates to AI Insights page", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    await page.locator('a[href="/ai"]').click();
    await page.waitForURL("**/ai");
    await expect(page.locator("h1")).toContainText("AI Insights");
  });

  test("navigates to Alerts page", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    await page.locator('a[href="/alerts"]').click();
    await page.waitForURL("**/alerts");
    await expect(page.locator("h1")).toContainText("Alert Feed");
  });

  test("navigates to Settings page", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    await page.locator('a[href="/settings"]').click();
    await page.waitForURL("**/settings");
    await expect(page.locator("h1")).toContainText("Settings");
  });
});
