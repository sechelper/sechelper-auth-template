# Web

React 19.2.7 frontend for the authentication shell. The browser uses the backend session Cookie and does not receive Client Secret or access token values.

The public frontend is built from this directory and served at `/`. Framework code lives under `src/framework`, project pages live under `src/business/<module>`, and `src/main.jsx` is the composition entry point. The independent management frontend lives in `web/admin/`, keeps its own Vite project and dependency lockfile, and is served at `/admin`.

Canonical setup and deployment instructions live under the repository `docs/` directory.
