# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: responsive-a11y.spec.ts >> Accessibility >> all pages have proper heading structure
- Location: e2e\responsive-a11y.spec.ts:32:7

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('h1')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('h1')

```

```yaml
- text: "{\"success\":true,\"data\":{\"app\":\"Cromatic Vision Optical Platform\",\"mode\":\"in-memory (no Docker required)\",\"status\":\"running\",\"version\":\"1.0.0-dev\"},\"message\":\"Welcome to the Cromatic Vision Optical DEV API\"}"
```

# Test source

```ts
  1  | import { test, expect } from "@playwright/test";
  2  | 
  3  | test.describe("Responsive Design", () => {
  4  |   test("renders sidebar collapsed on mobile", async ({ page }) => {
  5  |     await page.setViewportSize({ width: 375, height: 812 });
  6  |     await page.goto("/");
  7  |     const sidebar = page.locator("aside");
  8  |     await expect(sidebar).toBeVisible();
  9  |     // On mobile, sidebar should be narrow (w-16 = 64px)
  10 |     const box = await sidebar.boundingBox();
  11 |     expect(box!.width).toBeLessThanOrEqual(80);
  12 |   });
  13 | 
  14 |   test("renders sidebar expanded on desktop", async ({ page }) => {
  15 |     await page.setViewportSize({ width: 1440, height: 900 });
  16 |     await page.goto("/");
  17 |     const sidebar = page.locator("aside");
  18 |     await expect(sidebar).toBeVisible();
  19 |     // On desktop, sidebar should be wider (w-64 = 256px)
  20 |     const box = await sidebar.boundingBox();
  21 |     expect(box!.width).toBeGreaterThanOrEqual(200);
  22 |   });
  23 | 
  24 |   test("dashboard content is visible on mobile", async ({ page }) => {
  25 |     await page.setViewportSize({ width: 375, height: 812 });
  26 |     await page.goto("/");
  27 |     await expect(page.locator("h1")).toContainText("Engineering Intelligence");
  28 |   });
  29 | });
  30 | 
  31 | test.describe("Accessibility", () => {
  32 |   test("all pages have proper heading structure", async ({ page }) => {
  33 |     const pages = ["/", "/topology", "/incidents", "/health", "/ai", "/alerts", "/settings"];
  34 |     for (const path of pages) {
  35 |       await page.goto(path);
  36 |       const h1 = page.locator("h1");
> 37 |       await expect(h1).toBeVisible();
     |                        ^ Error: expect(locator).toBeVisible() failed
  38 |       const text = await h1.textContent();
  39 |       expect(text!.length).toBeGreaterThan(0);
  40 |     }
  41 |   });
  42 | 
  43 |   test("navigation links have accessible names", async ({ page }) => {
  44 |     await page.goto("/");
  45 |     const links = page.locator("aside a");
  46 |     const count = await links.count();
  47 |     expect(count).toBeGreaterThanOrEqual(7);
  48 |   });
  49 | 
  50 |   test("no broken images or missing resources", async ({ page }) => {
  51 |     const failedRequests: string[] = [];
  52 |     page.on("response", (response) => {
  53 |       if (response.status() >= 400 && !response.url().includes("favicon")) {
  54 |         failedRequests.push(`${response.status()} ${response.url()}`);
  55 |       }
  56 |     });
  57 |     await page.goto("/");
  58 |     await page.waitForLoadState("networkidle");
  59 |     expect(failedRequests).toHaveLength(0);
  60 |   });
  61 | });
  62 | 
  63 | test.describe("Performance", () => {
  64 |   test("home page loads within 5 seconds", async ({ page }) => {
  65 |     const start = Date.now();
  66 |     await page.goto("/");
  67 |     await page.locator("h1").waitFor();
  68 |     const loadTime = Date.now() - start;
  69 |     expect(loadTime).toBeLessThan(5000);
  70 |   });
  71 | 
  72 |   test("no console errors on page load", async ({ page }) => {
  73 |     const errors: string[] = [];
  74 |     page.on("console", (msg) => {
  75 |       if (msg.type() === "error") {
  76 |         errors.push(msg.text());
  77 |       }
  78 |     });
  79 |     await page.goto("/");
  80 |     await page.waitForLoadState("networkidle");
  81 |     // Filter out known non-critical errors (e.g., favicon)
  82 |     const criticalErrors = errors.filter(
  83 |       (e) => !e.includes("favicon") && !e.includes("404")
  84 |     );
  85 |     expect(criticalErrors).toHaveLength(0);
  86 |   });
  87 | });
  88 | 
```