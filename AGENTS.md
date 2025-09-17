# Repository Guidelines

## Project Structure & Module Organization
- Root Go app: `main.go`, `app.go`, `go.mod`, `wails.json`.
- Frontend (Svelte + TS + Vite): `frontend/` → source in `frontend/src`, assets in `frontend/src/assets`, built output in `frontend/dist` (embedded via `//go:embed`).
- Build artifacts: `build/` (created by Wails).

## Build, Test, and Development Commands
- Live app (Wails): `wails dev` — runs Vite dev server and desktop shell with hot reload.
- Build production app: `wails build` — bundles frontend and outputs binaries to `build/`.
- Frontend only:
  - Dev server: `npm -C frontend run dev`
  - Type checks: `npm -C frontend run check`
  - Production build: `npm -C frontend run build`
- Go formatting/vet: `go fmt ./... && go vet ./...`

## Coding Style & Naming Conventions
- Go: use `gofmt` defaults; exported identifiers use PascalCase with doc comments; packages are lowercase, short, no underscores.
- Svelte/TS: 2‑space indent; components PascalCase (`App.svelte`), modules kebab‑case (`main.ts`, `style.css`). Keep UI logic in Svelte files; put helpers in `frontend/src/` modules.
- Imports: prefer relative paths in frontend; group std/lib/external imports in Go.

## Testing Guidelines
- Go tests: place `*_test.go` alongside code; run with `go test ./...`.
- Frontend: static/type checks via `npm -C frontend run check`. No unit test runner is configured; if adding one, use Vitest and place tests as `src/**/*.spec.ts`.

## Commit & Pull Request Guidelines
- Commits: concise, imperative mood (e.g., "feat: add image viewer", "fix: guard nil ctx"). Keep changes scoped.
- PRs: include a clear description, linked issues, and steps to verify. For UI changes, add screenshots or a short clip. Note platform(s) tested.

## Security & Configuration Tips
- Go version: defined in `go.mod`. Node/Vite versions are pinned in `frontend/package.json`.
- Do not commit secrets. Configuration lives in `wails.json`; prefer runtime/env variables if adding credentials or endpoints.
