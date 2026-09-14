import { test, expect } from "@playwright/test";

const username = process.env.E2E_USERNAME;
const password = process.env.E2E_PASSWORD;

async function completeProviderLogin(page) {
  if (!username || !password) return false;
  const passwordField = page.locator('input[type="password"]');
  if (await passwordField.count()) {
    const userField = page.locator('input[type="text"], input[name*="user" i], input[name*="account" i]').first();
    await userField.fill(username);
    await passwordField.first().fill(password);
    await page.getByRole("button", { name: /登录|sign in|log in/i }).click();
  }
  const consent = page.getByRole("button", { name: /确认并继续|authorize|allow|continue/i });
  if (await consent.count()) await consent.click();
  return true;
}

test("public shell and unauthenticated API are available", async ({ page, request }) => {
  await page.goto("/");
  await expect(page.getByRole("heading", { name: "统一认证业务框架" })).toBeVisible();
  const session = await request.get("/v1/auth/session");
  expect(session.ok()).toBeTruthy();
  expect((await session.json()).authenticated).toBe(false);
  const authorization = await request.get("/v1/authorization/me");
  expect(authorization.status()).toBe(401);
});

test("metrics endpoint exposes auth and rate-limit counters", async ({ request }) => {
  const response = await request.get("/metrics");
  expect(response.ok()).toBeTruthy();
  const body = await response.text();
  expect(body).toContain("auth_login_total");
  expect(body).toContain("auth_rate_limited_total");
});

test("real test IdP login and logout", async ({ page }) => {
  test.skip(!username || !password, "Set E2E_USERNAME and E2E_PASSWORD for the real IdP flow.");
  await page.goto("/");
  await page.getByRole("button", { name: "登录" }).click();
  await completeProviderLogin(page);
  await page.waitForURL(/order-test\.sechelper\.com\/$/, { timeout: 30_000 });
  await expect(page.getByText("当前会话")).toBeVisible();
  await page.getByRole("button", { name: "退出登录" }).click();
  await expect(page.getByText("当前未登录。")).toBeVisible();
});

