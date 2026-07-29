# golang-develop example

这个目录是一个 Go 服务项目模板。详细说明请查看每个目录下的 README.md 文件。

## 设计原则

- `cmd` 做进程级装配，`app` 做应用级装配，`config` 描述运行契约。
- 不把 application-level wiring 放进 `cmd`：否则入口会直接依赖数据库、缓存、消息队列和业务服务，CLI surface、进程生命周期、应用依赖图会混在一起。
- `handler` 做协议适配，`service` 做业务用例，`repository` 做存储适配。
- 先让一个应用清楚可读，再考虑抽象复用。
- 单次使用的 option builder、parse helper 不抽函数；通用且复用的 helper 放 `pkg/util/...`。
- `configs/config.example.yaml` 或 `configs/<app>/config.example.yaml` 必须能支撑本地启动和 README 示例。
- 改完模板后至少跑 `go test ./...`、`go build ./cmd/app1`，并通过 CLI 路径手工验证。

## Go Directories

### `cmd/`

详情请参阅 [cmd/](cmd/)

### `internal/`

详情请参阅 [internal/](internal/)

### `pkg/`

详情请参阅 [pkg/](pkg/)
