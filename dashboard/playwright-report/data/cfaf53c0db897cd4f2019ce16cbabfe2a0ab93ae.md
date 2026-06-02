# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: alerts.spec.ts >> Alerts Page >> displays alert cards
- Location: e2e\alerts.spec.ts:12:7

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('text=CRITICAL: PostgreSQL pool exhausted')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('text=CRITICAL: PostgreSQL pool exhausted')

```

```yaml
- text: "{\"success\":false,\"error\":{\"code\":\"ERROR\",\"message\":\"Cannot GET /alerts\"}}"
```

# Test source

```ts
  1  | import { test, expect } from "@playwright/test";
  2  | 
  3  | test.describe("Alerts Page", () => {
  4  |   test.beforeEach(async ({ page }) => {
  5  |     await page.goto("/alerts");
  6  |   });
  7  | 
  8  |   test("renders page title", async ({ page }) => {
  9  |     await expect(page.locator("h1")).toContainText("Alert Feed");
  10 |   });
  11 | 
  12 |   test("displays alert cards", async ({ page }) => {
  13 |     await expect(
  14 |       page.locator("text=CRITICAL: PostgreSQL pool exhausted")
> 15 |     ).toBeVisible();
     |       ^ Error: expect(locator).toBeVisible() failed
  16 |     await expect(page.locator("text=MAJOR: API latency exceeds SLO")).toBeVisible();
  17 |     await expect(page.locator("text=WARNING: Redis cache degraded")).toBeVisible();
  18 |   });
  19 | 
  20 |   test("shows alert severity indicators", async ({ page }) => {
  21 |     // Critical alerts should be present
  22 |     const alerts = page.locator('[class*="border"]');
  23 |     await expect(alerts.first()).toBeVisible();
  24 |   });
  25 | });
  26 | 
```