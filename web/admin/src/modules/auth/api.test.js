import test from "node:test";
import assert from "node:assert/strict";
import { SESSION_REFRESH_LEEWAY_MS, sessionNeedsRefresh } from "./api.js";

test("admin session refresh starts within the shared leeway", () => {
  const now = Date.parse("2026-09-20T00:00:00Z");
  assert.equal(sessionNeedsRefresh(new Date(now + SESSION_REFRESH_LEEWAY_MS).toISOString(), now), true);
  assert.equal(sessionNeedsRefresh(new Date(now + SESSION_REFRESH_LEEWAY_MS + 1).toISOString(), now), false);
  assert.equal(sessionNeedsRefresh(new Date(now - 1).toISOString(), now), true);
  assert.equal(sessionNeedsRefresh("not-a-date", now), false);
});
