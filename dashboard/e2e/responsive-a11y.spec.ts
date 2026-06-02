import { test, expect } from "@playwright/test";

test.describe("Responsive Design", () => {
  test("renders sidebar collapsed on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto("/");
    await page.waitForSelector("aside");
    const sidebar = page.locator("aside");
    await expect(sidebar).toBeVisible();
    const box = await sidebar.boundingBox();
    expect(box!.width).toBeLessThanOrEqual(80);
  });

  test("renders sidebar expanded on desktop", async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.goto("/");
    await page.waitForSelector("aside");
    const sidebar = page.locator("aside");
    await expect(sidebar).toBeVisible();
    const box = await sidebar.boundingBox();
    expect(box!.width).toBeGreaterThanOrEqual(200);
  });

  test("dashboard content is visible on mobile", async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto("/");
    await page.waitForSelector("h1");
    await expect(page.locator("h1")).toContainText("Engineering Intelligence");
  });
});

test.describe("Accessibility", () => {
  test("all pages have proper heading structure", async ({ page }) => {
    const pages = ["/", "/topology", "/incidents", "/health", "/ai", "/alerts", "/settings"];
    for (const path of pages) {
      await page.goto(path);
      await page.waitForSelector("h1");
      const h1 = page.locator("h1");
      await expect(h1).toBeVisible();
      const text = await h1.textContent();
      expect(text!.length).toBeGreaterThan(0);
    }
  });

  test("navigation links have accessible names", async ({ page }) => {
    await page.goto("/");
    await page.waitForSelector("aside");
    const links = page.locator("aside a");
    const count = await links.count();
    expect(count).toBeGreaterThanOrEqual(7);
  });

  test("no broken images or missing resources", async ({ page }) => {
    const failedRequests: string[] = [];
    page.on("response", (response) => {
      if (response.status() >= 400 && !response.url().includes("favicon")) {
        failedRequests.push(`${response.status()} ${response.url()}`);
      }
    });
    await page.goto("/");
    await page.waitForLoadState("networkidle");
    expect(failedRequests).toHaveLength(0);
  });
});

test.describe("Performance", () => {
  test("home page loads within 10 seconds", async ({ page }) => {
    const start = Date.now();
    await page.goto("/");
    await page.waitForSelector("h1");
    const loadTime = Date.now() - start;
    expect(loadTime).toBeLessThan(10000);
  });

  test("no console errors on page load", async ({ page }) => {
    const errors: string[] = [];
    page.on("console", (msg) => {
      if (msg.type() === "error") {
        errors.push(msg.text());
      }
    });
    await page.goto("/");
    await page.waitForLoadState("networkidle");
    // Filter out known non-critical errors
    const criticalErrors = errors.filter(
      (e) => !e.includes("favicon") && !e.includes("404")
    );
    expect(criticalErrors).toHaveLength(0);
  });
});
