import { test, expect } from "@playwright/test";

test.describe("Topology Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/topology");
    await page.waitForSelector("h1");
  });

  test("renders page title and description", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("Infrastructure Topology");
    await expect(
      page.getByText("Component dependency graph")
    ).toBeVisible();
  });

  test("displays topology nodes", async ({ page }) => {
    // Wait for animations to complete (Framer Motion spring animations)
    await expect(page.getByText("Backend API")).toBeVisible({ timeout: 10000 });
    await expect(page.getByText("Auth Service")).toBeVisible();
    await expect(page.getByText("PostgreSQL Primary")).toBeVisible();
    await expect(page.getByText("Redis Cache")).toBeVisible();
    await expect(page.getByText("NGINX Proxy")).toBeVisible();
  });

  test("topology nodes are interactive (clickable)", async ({ page }) => {
    const node = page.getByText("Backend API");
    await expect(node).toBeVisible();
    await node.click();
    await expect(node).toBeVisible();
  });
});
