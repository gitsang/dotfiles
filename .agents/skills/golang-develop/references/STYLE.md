<!-- MarkdownTitleNumber auto -->

# 代码风格

## 1. 实例构造：Functional Option

**所有可配置组件使用 Functional Option 模式**:

```go
type AgentBridge struct {
    client    sessionClient
    timeout   time.Duration
    retryMax  int
}

type OptionFunc func(*AgentBridge)

func WithClient(client sessionClient) OptionFunc {
    return func(target *AgentBridge) {
        target.client = client
    }
}

func WithTimeout(timeout time.Duration) OptionFunc {
    return func(target *AgentBridge) {
        target.timeout = timeout
    }
}

func New(optfs ...OptionFunc) *AgentBridge {
    connector := &AgentBridge{
        timeout:  30 * time.Second, // 默认值
        retryMax: 3,
    }
    for _, apply := range optfs {
        if apply == nil {
            continue
        }
        apply(connector)
    }
    return connector
}

// 使用
bridge := New(
    WithClient(client),
    WithTimeout(60*time.Second),
)
```

## 2. 日志规范

**使用 `log/slog` 做结构化日志**:

```go
import "log/slog"

// 1. 通过 logger.With(...) 挂稳定字段
logger := slog.New(handler).With(
    slog.String("service", "user-api"),
    slog.String("env", "production"),
)

// 2. 函数内逐步增加日志内容
func ProcessOrder(ctx context.Context, orderID string) {
    logger := logger.With(slog.String("orderID", orderID))

    // ... 业务逻辑

    // 3. 尽量只在函数结束（defer）时输出
    defer func() {
        logger.Debug("process completed", slog.String("status", status))
    }()
}
```

**日志原则**:

| 原则           | 说明                         |
| -------------- | ---------------------------- |
| 减少输出数量   | 不要在核心逻辑里堆噪音       |
| 增加单次信息量 | 一次日志包含完整上下文       |
| defer 输出     | 短生命周期函数只在结束时输出 |
| With(...) 累积 | 函数运行中逐步增加字段       |

## 3. Tracing

**多入口应用需要统一 trace 机制，将请求、任务或消息的日志关联起来**:

**设计核心**（参考 yauth-middleware/pkg/contextx）:

1. **入口提取 traceId**: HTTP 从 `X-Trace-Id` header 提取，消息/任务从元数据提取，存入 context
2. **类型安全**: 使用 Go 1.18+ 泛型 GetValueAs[T]
3. **线程安全**: sync.Map 包装（SafeContext）
4. **跨边界传播**: HTTP 场景同时设置 request/response header，消息或任务场景写入元数据

**实现示例**:

```go
// 1. context value helpers (util/context)
func GetValueAs[T any](ctx context.Context, key any) (T, error) {
    raw := ctx.Value(key)
    if raw == nil {
        return *new(T), NewNotFoundError(key)
    }
    value, ok := raw.(T)
    if !ok {
        return *new(T), NewIncorrectTypeError(key, new(T))
    }
    return value, nil
}

// 2. trace key
const TraceIdKey = "traceId"

// 3. HTTP trace middleware（HTTP API 场景）
func TraceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从 header 提取或生成
        traceId := r.Header.Get("X-Trace-Id")
        if traceId == "" {
            traceId = generateTraceId() // e.g. xid.New().String()
        }

        // 存入 context
        ctx := context.WithValue(r.Context(), TraceIdKey, traceId)

        // 设置 header（request + response）
        r.Header.Set("X-Trace-Id", traceId)
        w.Header().Set("X-Trace-Id", traceId)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 4. handler 中使用
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    traceId, _ := context.GetValueAs[string](r.Context(), TraceIdKey)
    logger := slog.New(handler).With(slog.String("traceId", traceId))

    logger.Info("request received", slog.String("path", r.URL.Path))
}
```

## 4. 依赖注入

### 4.1 核心原则

**不使用 DI 框架（如 wire），手动注入让依赖关系清晰可见**

参考: [DI Containers are Code Polluters](https://www.yegor256.com/2014/10/03/di-containers-are-evil.html) by Yegor Bugayenko

**为何不使用 wire 等框架**:

| 问题         | 说明                                         |
| ------------ | -------------------------------------------- |
| 隐藏依赖关系 | DI 容器把依赖关系藏在配置/注解中，代码不可读 |
| 增加复杂度   | 框架本身是额外负担，需要学习、维护、调试     |
| 降低可测试性 | 自动生成的代码难以 mock，测试复杂度增加      |
| 违反封装原则 | 依赖变成"魔法"，破坏了显式组合的原则         |

### 4.2 正确做法: 手动 Functional Options

```go
// 定义结构和选项函数
type UserService struct {
    db        DB
    cache     Cache
    logger    *slog.Logger
    timeout   time.Duration
}

type OptionFunc func(*UserService)

func WithDB(db DB) OptionFunc {
    return func(s *UserService) { s.db = db }
}

func WithCache(cache Cache) OptionFunc {
    return func(s *UserService) { s.cache = cache }
}

func WithLogger(logger *slog.Logger) OptionFunc {
    return func(s *UserService) { s.logger = logger }
}

func WithTimeout(timeout time.Duration) OptionFunc {
    return func(s *UserService) { s.timeout = timeout }
}

// 构造函数: 设置默认值 + 应用选项
func NewUserService(optfs ...OptionFunc) *UserService {
    service := &UserService{
        timeout: 30 * time.Second,  // 默认值
    }
    for _, apply := range optfs {
        if apply != nil {
            apply(service)
        }
    }
    return service
}
```

### 4.3 手动装配分层

依赖注入不是“所有构造代码都放到 `main.go`”。本模板按层分开：

- `cmd/<app>/main.go` 做 **process-level wiring**：Cobra、flag/env/config loading、早期 logger、signal、exit code、调用应用生命周期。
- `internal/<app>/app` 做 **application-level wiring**：server、handler、service、repository、cache、mq、worker 的构造与关闭。
- 业务逻辑放在 `handler` / `service` / `repository` / `worker` 等包，不放 `cmd` 或 `app`。

#### 4.3.1 `cmd` 只做进程级装配

```go
// cmd/app/main.go - 进程入口，调用应用生命周期
func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func Run() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    cfg := loadConfig(rootFlags.ConfigFile)
    logger := newLogger(cfg.Log)

    application, err := app.New(ctx, cfg, logger)
    if err != nil {
        logger.Error("create application failed", slog.Any("err", err))
        os.Exit(1)
    }
    defer application.Shutdown(context.Background())

    setupGracefulShutdown(cancel, application)

    if err := application.Start(); err != nil {
        logger.Error("serve failed", slog.Any("err", err))
        os.Exit(1)
    }
}
```

#### 4.3.2 `internal/<app>/app` 做应用级装配

```go
// internal/app/app.go - 应用运行体的组合根
func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
    db, err := repository.NewDB(cfg.Data.Database)
    if err != nil {
        return nil, err
    }

    cache, err := cachex.New(cfg.Data.Cache)
    if err != nil {
        return nil, errors.Join(db.Close(), err)
    }

    userService := service.NewUserService(
        service.WithDB(db),
        service.WithCache(cache),
        service.WithLogger(logger),
    )

    handler := api.NewHandler(
        api.WithUserService(userService),
        api.WithLogger(logger),
    )

    server := &http.Server{
        Addr:    cfg.Server.HTTP.Addr,
        Handler: handler,
    }

    return &App{
        db:     db,
        cache:  cache,
        server: server,
        logger: logger,
    }, nil
}
```

#### 4.3.3 为什么不在 `cmd` 做 application-level wiring

- `cmd` 是用户直接接触的 CLI / process surface，应保持稳定、薄、可预测。
- 如果 `cmd` 直接 import database、cache、mq、handler、service，会同时承担入口、配置、依赖图、生命周期，边界会变成大杂烩。
- `internal/<app>/app.New` 可以单独测试装配失败、资源关闭和生命周期；放进 `cmd` 后通常只能通过真实进程或 CLI 间接测。
- 多 binary 项目中，`cmd/<app>` 应只表达每个 binary 的启动方式差异；应用内部依赖图仍由对应的 `internal/<app>/app` 管理。
- Dapr、Kubernetes 等大型项目把 `app` 放在 `cmd/<binary>/app`，本质也是 binary composition root；模板不应照搬路径，而应保留清晰的 process-level / application-level 边界。
