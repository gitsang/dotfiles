<!-- MarkdownTitleNumber auto -->

# 基础库

优先使用 **ykit** 仓库的基础组件，避免造轮子。

```
gitcode.yealink.com/server/devops/yops/ykit
```

## 1. ykit 包一览

| 包         | 用途                         | 初始化要求 |
| ---------- | ---------------------------- | ---------- |
| `util`     | 通用工具函数                 | 无需初始化 |
| `log`      | 结构化日志 (slog)            | 需初始化   |
| `cache`    | 缓存接口 (Redis/Memory)      | 需初始化   |
| `configer` | 配置管理 (viper)             | 需初始化   |
| `mq`       | 消息队列 (Producer/Consumer) | 需初始化   |
| `lock`     | 分布式锁 (Redis)             | 需初始化   |

## 2. util - 通用工具

**何时使用**: 字符串处理、类型转换、错误处理、优雅关闭等辅助场景。

**核心子包**:

```go
import "gitcode.yealink.com/server/devops/yops/ykit/util/candy"
import "gitcode.yealink.com/server/devops/yops/ykit/util/context"
import "gitcode.yealink.com/server/devops/yops/ykit/util/runtime"

// candy: 错误处理助手
result := candy.Must(doSomething())  // 失败则 panic
result := candy.Ignore(doSomething()) // 忽略错误

// context: 类型安全的 context 值获取
value, err := context.GetValueAs[T](ctx, key)

// runtime: 优雅关闭
runtime.SetupGracefulShutdown(ctx, func(sig os.Signal) {
    // 清理资源
    server.Shutdown(ctx)
})
```

## 3. log - 结构化日志

**何时使用**: 全局日志输出，支持多输出端（stdout + file）、多格式（console + json）。

**核心设计**: 基于 `log/slog`，Handler 可组合。

```go
import "gitcode.yealink.com/server/devops/yops/ykit/log"
import "gitcode.yealink.com/server/devops/yops/ykit/log/ylog"
import "log/slog"

handler := log.NewHandler(log.HandlerOptions{
    Format:     "console",   // console, json, text（生产环境通常首选 json，以便日志采集和解析）
    Level:      "info",      // debug, info, warn, error
    Color:      true,
    Writers:    []io.Writer{os.Stdout},
    Verbosity:  1,           // 0=无源码, 1=简短源码, 2=完整函数名, 3=Source结构, 4=完整路径
    CallerSkip: 0,
})
logger := slog.New(handler)

logger.Info("server started", slog.String("port", "8080"))
logger.Error("failed to connect", slog.Any("err", err))
```

## 4. cache - 缓存接口

**何时使用**: 需要缓存高频访问数据以减少数据库/服务压力。

**核心接口**:

```go
import "gitcode.yealink.com/server/devops/yops/ykit/cache"
import "gitcode.yealink.com/server/devops/yops/ykit/cache/pkg/redisc"

// Redis 缓存
c, err := redisc.New(&redisc.Option{
    Addrs:    []string{"localhost:6379"},
    Password: "",
    DB:       0,
})

// 带命名空间隔离（推荐）
c = c.WithNamespaces("myapp", "users")

// 基础操作
value, err := c.Get(ctx, "key")        // 无过期
err = c.Set(ctx, "key", "value")       // 无过期
value, err := c.GetEx(ctx, "key", ttl) // 带过期检查
err = c.SetEx(ctx, "key", "value", ttl) // 带过期设置
err = c.Del(ctx, "key")
ttl, err := c.Ttl(ctx, "key")
```

**requester - 防击穿缓存**:

```go
import "gitcode.yealink.com/server/devops/yops/ykit/cache/requester"

// 分布式场景使用 requester 防止缓存击穿
cr := requester.NewCacheRequester(
    requester.WithCache(cacheClient),
    requester.WithDistributedLock(lockClient),
    requester.WithTtl(10*time.Second),
    requester.WithMaxTtl(1*time.Minute),
)

// 调用：先查缓存，未命中则加锁执行 f，结果存缓存
resp, err := cr.RequestWithCache(ctx, "key", func() ([]byte, error) {
    return fetchFromDatabase()
})
```

## 5. configer - 配置管理

**何时使用**: 应用启动时加载配置，支持配置文件、环境变量、命令行参数。

**优先级**: 命令行参数 > 环境变量 > 配置文件 > 默认值

```go
import "gitcode.yealink.com/server/devops/yops/ykit/configer"

type Config struct {
    Server ServerConfig `yaml:"server"`
    DB     DBConfig     `yaml:"db"`
}

type ServerConfig struct {
    Port int    `yaml:"port" default:"8080" env:"SERVER_PORT" flag:"port" usage:"server port"`
    Host string `yaml:"host" default:"localhost" env:"SERVER_HOST"`
}

// 初始化 configer
cfger := configer.New(
    configer.WithTemplate(new(Config)),
    configer.WithEnvBind(
        configer.WithEnvPrefix("MYAPP"),
        configer.WithEnvDelim("_"),
    ),
    configer.WithFlagBind(
        configer.WithCommand(rootCmd),
        configer.WithFlagPrefix(""),
        configer.WithFlagDelim("-"),
    ),
)

// 加载配置
var cfg Config
err := cfger.Load(&cfg, "config.yaml", "config.local.yaml")
```

**环境变量示例**:

```bash
MYAPP_SERVER_PORT=9090      # 覆盖 server.port
MYAPP_SERVER_HOST=0.0.0.0   # 覆盖 server.host
```

## 6. mq - 消息队列

**何时使用**: 异步处理、事件驱动架构、服务解耦。

**核心接口**:

```go
import "gitcode.yealink.com/server/devops/yops/ykit/mq"

// 消息结构
msg := &mq.Message{
    Topic:   "order.created",
    Labels:  []string{"urgent"},
    Content: []byte(orderJSON),
}

// Producer 接口
type Producer interface {
    Connect() error
    Publish(ctx context.Context, msgs ...*Message) error
    Close() error
}

// Consumer 接口
type Consumer interface {
    Connect() error
    Subscribe(topicPattern string, labels []string, callback CallbackFunc) error
    Close() error
}

// 使用示例（具体驱动见 mq/example）
producer.Publish(ctx, msg)

consumer.Subscribe("order.*", nil, func(ctx context.Context, msgs ...*mq.MessageExt) error {
    for _, msg := range msgs {
        processOrder(msg.Content)
    }
    return nil
})
```

## 7. lock - 分布式锁

**何时使用**: 分布式系统中防止并发冲突、资源竞争。

```go
import "gitcode.yealink.com/server/devops/yops/ykit/lock"
import "gitcode.yealink.com/server/devops/yops/ykit/lock/rdslock"

// Redis 分布式锁
locker, err := rdslock.NewRedisLock(
    rdslock.WithClientOption(&rdslock.ClientOption{
        Addrs:      []string{"localhost:6379"},
        Password:   "",
        DB:         0,
        Namespaces: []string{"myapp"},
    }),
)

// 带命名空间隔离
locker = locker.WithNamespaces("resource")

// 获取锁
acquired, err := locker.Acquire(ctx, "user:123", 30*time.Second)
if !acquired {
    return errors.New("lock busy")
}
defer locker.Release(ctx, "user:123")

// 续期
locker.Renew(ctx, "user:123", 30*time.Second)
```
