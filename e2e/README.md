# Browser E2E

Install and run:

```bash
cd e2e
npm install
npx playwright install chromium
npm test
```

The public tests require only `E2E_BASE_URL`. The real IdP test is enabled only when `E2E_USERNAME` and `E2E_PASSWORD` are supplied through the test runner environment; never commit those values or a storage state file. Use `E2E_HEADED=1` for headed debugging and `E2E_BROWSER_CHANNEL=chrome` when a system Chrome is already available.

CI uses `.github/workflows/e2e.yml`; configure `E2E_BASE_URL` as an Actions variable and provide `E2E_USERNAME`/`E2E_PASSWORD` as Actions secrets. The real IdP test remains skipped when credentials are absent, so public smoke tests can run safely on forked builds.

