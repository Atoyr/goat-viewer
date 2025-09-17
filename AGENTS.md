# Repository Guidelines

## Project Structure & Module Organization
- `src/`: SvelteKit app (routes, UI). Key files: `routes/+page.svelte`, `routes/+layout.ts`, `app.html`.
- `static/`: Public assets served as-is (icons, images).
- `src-tauri/`: Tauri (Rust) backend. Important: `src/lib.rs` (commands), `src/main.rs` (entry), `tauri.conf.json` (app config), `Cargo.toml`.
- Build output: `build/` (frontend) consumed by Tauri per `tauri.conf.json`.

## Build, Test, and Development Commands
- `pnpm tauri dev`: Run desktop app in dev (Vite + Tauri).
- `pnpm tauri build`: Build installers/binaries for your platform.
- `pnpm dev`: Frontend-only Vite dev server (no Tauri shell).
- `pnpm build`: Build SvelteKit frontend to `build/`.
- `pnpm check`: Type-check Svelte/TS via `svelte-check`.
- Android (optional): `pnpm tauri android dev` (see README for init).

## Coding Style & Naming Conventions
- TypeScript/Svelte: 2-space indent, semicolons optional; prefer explicit types at boundaries. Keep `tsconfig.json` strict settings.
- SvelteKit routing: use `+page.svelte`, `+layout.ts` and colocate logic/styles with components.
- Components: `PascalCase.svelte`. Utility modules: `camelCase.ts`.
- Rust (Tauri): follow `rustfmt` defaults; module/func names `snake_case`, types `PascalCase`. Keep Tauri commands in `src-tauri/src/lib.rs` with `#[tauri::command]`.

## Testing Guidelines
- No test runner is configured yet. Use `pnpm check` for static checks.
- If adding tests, prefer Vitest for TS and Playwright for E2E; place unit tests next to sources as `*.test.ts`.
- Keep small, pure functions testable; isolate Tauri calls behind thin wrappers.

## Commit & Pull Request Guidelines
- Commits: concise, imperative subject. Prefer Conventional Commits (e.g., `feat:`, `fix:`, `chore:`).
- PRs: include summary, linked issues, platform(s) tested (Windows/macOS/Linux/Android), and screenshots for UI changes. Describe steps to verify.
- Keep PRs small and focused; update docs when behavior or commands change.

## Security & Configuration Tips
- Tauri config: review `src-tauri/tauri.conf.json` before release (icons, `identifier`, CSP). Avoid enabling dev-only options in production.
- Capabilities: use `src-tauri/capabilities` to scope APIs.
- Avoid remote code/URLs unless explicitly allowed; prefer bundling assets in `static/`.

