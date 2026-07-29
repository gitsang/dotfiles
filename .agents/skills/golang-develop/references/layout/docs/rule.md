# Coding Rules

## Keep assembly code readable

- Keep dependency assembly in the app assembly file unless there is real reuse or a clear domain boundary.
- `cmd` owns process-level wiring only; `internal/app` or `internal/<app>/app` owns application-level wiring.
- Do not initialize database, cache, message queues, HTTP routes, services, repositories, or workers directly in `cmd`.
- Do not add a separate `dependencies.go` just to move one app's construction code out of `app.go`.
- Do not extract one-off option builders or parse helpers when the extracted function only hides a short, local flow.
- Prefer seeing `strings.ToLower(strings.TrimSpace(driver))` at the decision point over a tiny wrapper such as `normalizeDriver`.
- Extract helpers when they name a reusable operation or remove meaningful duplication across call sites.

## Project layout

- For a single-app service, omit the app-name layer: use `cmd/main.go`, `internal/config`, and `internal/app`.
- For multiple apps, use `cmd/<app>`, `internal/<app>`, and `configs/<app>`.
- Put protobuf API contracts under `api/proto/<version>` and generate public API packages under `pkg/api/<version>`.
- Keep proto-to-HTTP generation declarative: `google.api.http` annotations live in `.proto`, `buf.gen.yaml` owns generator output paths, and runtime registration/handlers stay in `internal/<app>/handler`.
- Use `handler` for protocol adapters, `service` for application use cases, and `repository` for persistence adapters.
- Keep all config files under `configs/`; ignore real config files and commit only example configs.
- `make dev` should forward CLI args after `--`, so developers can run `APP=app1 make dev -- -c ./configs/app1/config.yml`.
- For forwarded long flags, prefer space-separated values (`--flag value`) because make treats `--flag=value` as a goal-like token.

## Shared utilities

- Put broadly reusable string helpers under `pkg/util/stringx`.
- Keep generic helpers small and behavior-focused. Example: `TrimSplit` splits, trims each item, and drops empty items.
