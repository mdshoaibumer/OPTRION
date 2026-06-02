import { test, expect } from "@playwright/test";

test.describe("Settings Page", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto("/settings");
    await page.waitForSelector("h1");
  });

  test("renders page title", async ({ page }) => {
    await expect(page.locator("h1")).toContainText("Settings");
    await expect(
      page.getByText("Configure your Optrion dashboard connection")
    ).toBeVisible();
  });

  test("displays API connection form", async ({ page }) => {
    await expect(page.getByText("API Connection")).toBeVisible();
    await expect(page.getByText("API URL")).toBeVisible();
    await expect(page.getByText("API Key", { exact: true })).toBeVisible();
  });

  test("API URL field has default value", async ({ page }) => {
    const urlInput = page.locator('input[placeholder="http://localhost:8080"]');
    await expect(urlInput).toHaveValue("http://localhost:8080");
  });

  test("can enter API key and save", async ({ page }) => {
    const apiKeyInput = page.locator('input[type="password"]');
    await apiKeyInput.fill("test-api-key-12345");
    await expect(apiKeyInput).toHaveValue("test-api-key-12345");

    // Click save button
    await page.getByRole("button", { name: /save settings/i }).click();

    // Should show success feedback
    await expect(page.getByText("Saved!")).toBeVisible();
  });

  test("persists settings to localStorage", async ({ page }) => {
    const urlInput = page.locator('input[placeholder="http://localhost:8080"]');
    const apiKeyInput = page.locator('input[type="password"]');

    await urlInput.fill("http://api.optrion.io:9090");
    await apiKeyInput.fill("my-secret-key");
    await page.getByRole("button", { name: /save settings/i }).click();

    // Verify localStorage was set
    const storedUrl = await page.evaluate(() =>
      localStorage.getItem("optrion_api_url")
    );
    const storedKey = await page.evaluate(() =>
      localStorage.getItem("optrion_api_key")
    );

    expect(storedUrl).toBe("http://api.optrion.io:9090");
    expect(storedKey).toBe("my-secret-key");
  });
});
