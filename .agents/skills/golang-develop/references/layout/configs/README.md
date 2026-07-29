# `configs/`

`configs/` 放可运行的示例配置文件。它的目标是让新项目能直接启动，并说明配置字段怎么写。

## 配置从哪里来

配置来源按这个思路组织：

1. `internal/config/config.go` 或 `internal/<app>/config/config.go` 的默认值提供完整 schema。
2. `configs/config.example.yaml` 覆盖单应用本地运行需要显式展示的值；多应用使用 `configs/<app>/config.example.yaml`。
3. 环境变量用于部署环境覆盖，例如地址、密码、外部服务端点。
4. 命令行 `--config.*` 用于临时调试和手工 QA。

当前示例：

```bash
go run ./cmd/app1 --config-file configs/app1/config.example.yaml
go run ./cmd/app1 --config-file configs/app1/config.example.yaml --config.server.http.addr=127.0.0.1:0
APP1_DATA_CACHE_DRIVER=memory go run ./cmd/app1 --config-file configs/app1/config.example.yaml
APP=app1 make dev -- -c ./configs/app1/config.example.yaml
cp configs/app1/config.example.yaml configs/app1/config.yml
APP=app1 make dev
```

通过 `make dev -- ...` 透传参数时，长 flag 建议使用空格分隔值，例如 `--config.server.http.addr 127.0.0.1:0`。

## 配置文件应该怎么放

- 所有配置文件都放到 `configs/` 下，不使用 `conf/`、项目根目录或其它临时目录。
- 单应用项目使用 `configs/config.example.yaml`。
- 多应用项目使用 `configs/<app>/config.example.yaml`，例如当前模板的 `configs/app1/config.example.yaml`。
- `make dev` 不内置 `--config-file`，默认依赖 CLI 的 `configs/<app>/config.yml`；这个真实配置被 `.gitignore` 忽略，开发者可从 example 复制。
- `.gitignore` 默认忽略 `configs/` 下所有 `*.yaml`、`*.yml`、`*.json`、`*.toml` 真实配置文件，只放行 `*example.*` 文件。

## `config.example.yaml` 应该放什么

应该放：

- 本地可运行的 server 地址。
- stdout/stderr/file 等日志样例。
- memory cache/mq 这类不依赖外部服务的默认 driver。
- 非敏感占位值。

不应该放：

- 真实密码、token、证书。
- 生产环境地址。
- 和 `config.go` 不一致的废弃字段。

## 新增配置字段的步骤

1. 先在 `internal/config/config.go` 或 `internal/<app>/config/config.go` 加字段和默认值。
2. 如果本地启动需要展示，就同步 `configs/config.example.yaml` 或 `configs/<app>/config.example.yaml`。
3. 在 `internal/app`、`internal/<app>/app` 或业务包中消费该字段。
4. 跑 `go run ./cmd/<app> --help`（单应用为 `go run ./cmd --help`），确认 `--config.*` flag 能展示出来。
5. 用一次 CLI 覆盖验证字段路径是否正确。

## 命名规则

- YAML 使用结构体路径的小驼峰形式：`server.http.addr`、`data.cache.driver`。
- CLI 使用 `--config.` 前缀：`--config.server.http.addr`。
- env 使用应用前缀和 `_`：`APP1_SERVER_HTTP_ADDR`。

示例配置必须服务于 README 的启动命令和 `make dev`；如果命令不能用，优先修正配置或 Makefile。
