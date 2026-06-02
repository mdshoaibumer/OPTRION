# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: navigation.spec.ts >> Navigation >> navigates to AI Insights page
- Location: e2e\navigation.spec.ts:32:7

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('a[href="/ai"]')

```

# Page snapshot

```yaml
- generic [ref=e2]: "{\"success\":true,\"data\":{\"app\":\"Cromatic Vision Optical Platform\",\"mode\":\"in-memory (no Docker required)\",\"status\":\"running\",\"version\":\"1.0.0-dev\"},\"message\":\"Welcome to the Cromatic Vision Optical DEV API\"}"
```

# Test source

```ts
  1  | import { test, expect } from "@playwright/test";
  2  | 
  3  | test.describe("Navigation", () => {
  4  |   test("sidebar renders all navigation items", async ({ page }) => {
  5  |     await page.goto("/");
  6  |     const nav = page.locator("aside");
  7  |     await expect(nav).toBeVisible();
  8  |     await expect(nav.locator("text=OPTRION")).toBeVisible();
  9  |   });
  10 | 
  11 |   test("navigates to Topology page", async ({ page }) => {
  12 |     await page.goto("/");
  13 |     await page.click('a[href="/topology"]');
  14 |     await expect(page).toHaveURL("/topology");
  15 |     await expect(page.locator("h1")).toContainText("Infrastructure Topology");
  16 |   });
  17 | 
  18 |   test("navigates to Incidents page", async ({ page }) => {
  19 |     await page.goto("/");
  20 |     await page.click('a[href="/incidents"]');
  21 |     await expect(page).toHaveURL("/incidents");
  22 |     await expect(page.locator("h1")).toContainText("Incident War Room");
  23 |   });
  24 | 
  25 |   test("navigates to Health page", async ({ page }) => {
  26 |     await page.goto("/");
  27 |     await page.click('a[href="/health"]');
  28 |     await expect(page).toHaveURL("/health");
  29 |     await expect(page.locator("h1")).toContainText("Component Health");
  30 |   });
  31 | 
  32 |   test("navigates to AI Insights page", async ({ page }) => {
  33 |     await page.goto("/");
> 34 |     await page.click('a[href="/ai"]');
     |                ^ Error: page.click: Test timeout of 30000ms exceeded.
  35 |     await expect(page).toHaveURL("/ai");
  36 |     await expect(page.locator("h1")).toContainText("AI Root Cause Analysis");
  37 |   });
  38 | 
  39 |   test("navigates to Alerts page", async ({ page }) => {
  40 |     await page.goto("/");
  41 |     await page.click('a[href="/alerts"]');
  42 |     await expect(page).toHaveURL("/alerts");
  43 |     await expect(page.locator("h1")).toContainText("Alert Feed");
  44 |   });
  45 | 
  46 |   test("navigates to Settings page", async ({ page }) => {
  47 |     await page.goto("/");
  48 |     await page.click('a[href="/settings"]');
  49 |     await expect(page).toHaveURL("/settings");
  50 |     await expect(page.locator("h1")).toContainText("Settings");
  51 |   });
  52 | });
  53 | 
```