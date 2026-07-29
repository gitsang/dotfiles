# `internal/`

`internal/` 放应用私有代码。Go 会限制外部模块 import `internal` 下的包，所以这里适合放服务自己的实现细节。

只有一个应用时可以省略应用名这一层：

```text
internal/
  config/
  app/
```

多个应用时再使用应用名分组。当前模板是多应用预留结构：

```text
internal/
  app1/
    config/
    app/
```

以后新增第二个应用时，切换到 `internal/<app>/...`，不要把多个应用的私有代码混在同一个包里。

## 先写 `config/`

`internal/config/config.go` 或 `internal/<app>/config/config.go` 是运行契约，应该最先写。

应该放：

- `Config` 结构体。
- 默认值 struct tag。
- 日志、server、database、cache、mq、外部依赖等配置分组。
- `NewConfiger(rootCmd)`，集中声明 env 和 flag 绑定规则。

当前模板规则：

- env 前缀：`APP1`
- env 分隔符：`_`
- flag 前缀：`config`
- flag 分隔符：`.`

映射示例：

| 配置字段 | YAML | 环境变量 | CLI flag |
| --- | --- | --- | --- |
| `Server.Http.Addr` | `server.http.addr` | `APP1_SERVER_HTTP_ADDR` | `--config.server.http.addr` |
| `Data.Cache.Driver` | `data.cache.driver` | `APP1_DATA_CACHE_DRIVER` | `--config.data.cache.driver` |
| `Data.Mq.KeepaliveInterval` | `data.mq.keepaliveInterval` | `APP1_DATA_MQ_KEEPALIVEINTERVAL` | `--config.data.mq.keepaliveinterval` |

为什么先写 config：它决定 `configs/` 示例、`cmd` flags、`app` 装配和测试用例的输入。

## 再写 `app/`

`internal/app/app.go` 或 `internal/<app>/app/app.go` 是应用装配和生命周期文件，也就是 application-level wiring 的位置。

应该放：

- `type App struct`：保存应用拥有的资源，例如 cache、mq、handler server、logger、cancel、shutdown 状态。
- `func New(ctx, cfg, logger) (*App, error)`：按配置创建基础设施，失败时关闭已创建资源。
- `func (a *App) Start() error`：阻塞运行服务，例如 `http.Server.ListenAndServe()`。
- `func (a *App) Shutdown(ctx) error`：幂等关闭 server、context、cache、mq、数据库等资源。

不要放：

- 不要把同一个应用的一次性装配拆到 `dependencies.go`。
- 不要提前引入 DI 框架。
- 不要为了隐藏几行代码抽 `newXOptions`、`parseOptionalDuration` 这类只用一次的 helper。

为什么集中装配：读 `app.go` 就能知道这个服务启动时依赖什么、失败时怎么清理、退出时关什么。

为什么不放在 `cmd`：`cmd` 只表达进程入口和启动方式。若在 `cmd` 中直接创建数据库、cache、mq、routes、service，入口会同时承担 CLI、配置、依赖图和生命周期，`app.New` 的装配失败与资源关闭也无法脱离真实进程独立测试。

## 业务代码放哪里

按需要添加目录，保持应用私有：

- `internal/<app>/handler`：HTTP/gRPC handler、routes、中间件，只做协议适配。
- `internal/<app>/service`：业务用例和领域流程，编排 repository/client/domain。
- `internal/<app>/repository`：存储接口和实现，对 service 屏蔽存储细节。
- `internal/<app>/worker`：后台消费、定时任务。

单应用项目对应 `internal/handler`、`internal/service`、`internal/repository`、`internal/worker`。

没有真实需求时不要先建空抽象层。先把一个用例写清楚，再按职责拆目录。

## 测试建议

- 配置映射、装配失败、资源关闭要测 `internal/app` 或 `internal/<app>/app`。
- 业务规则测对应 `service` 包。
- HTTP/gRPC surface 测 handler，不要只测内部 helper。
