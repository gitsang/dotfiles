---
name: golang-develop
description: >-
  Use when developing Go applications - guides project structure, base components (ykit), code patterns (Functional Options,
  structured logging, trace propagation), error handling, DI, testing, database operations, and optional HTTP API design.
  Triggers: Go application development, backend service, CLI or worker service, API development, microservice.
---

<!-- MarkdownTitleNumber auto -->

# Golang 开发指南

Go 应用开发的标准化指南，适用于后端服务、CLI、worker、HTTP API 等场景，用于构建健壮、高效且可维护的应用程序。

## 1. 何时使用

- 编写新的 Go 代码
- 审查或审计现有的 Go 代码
- 重构 Go 代码
- 关于 Go 项目结构、错误处理、并发或测试的任何问题

## 2. 核心原则

### 2.1 接口设计

#### 2.1.1 封装原则

定义契约而不是定义实现，服务应该 **对上层使用者屏蔽底层实现细节**。

通过定义清晰的接口向上暴露能力，或向下需求依赖。

调用方只关心：

- 能做什么（方法签名、输入输出）
- 需要什么（参数、前置条件）
- 得到什么（返回值、副作用、错误类型）

而不关心：

- 数据存在 MySQL 还是 PostgreSQL
- 缓存用 Redis 还是 Memcached
- 内部是同步调用还是异步消息队列
- 具体算法实现

这一原则贯穿本指南的所有内容：

- **定义契约**：实现库应该被实例化过程引用，而不是被使用者引用。
- **依赖注入**: 注入接口，而不是结构体。
- **错误处理**: 上层服务应该处理底层错误，而不是返回底层错误。
- **模型转换**：上层服务应该对底层模型进行转换，而不是直接使用。

#### 2.1.2 最小化接口

定义最小化的接口，不要定义一个巨大的接口，应该将接口拆分为多个小接口，如：

```go
type WriteableAgent interface {}
type AgentSessionManager interface {}
type QuestionableAgent interface {}
type AgentCore interface {}
```

而不是定义一个巨大的 `type Agent interface`，
巨大的接口可能迫使实现者编写 `panic("not implemented")` 或 `return errors.New("not implemented")` 这样的实现，
拆分接口实现者可以不必一次性实现全部接口，
而调用者也可以根据接口实现情况判断如何调用。

#### 2.1.3 隐式实现接口

接口属于使用它们的包，而非实现它们的包。接口是隐式的被实现的，实现者不应该知道我实现了哪个接口。

除非你正在定义标准接口库，否则在接口被使用的地方定义它们。

```
// processor/processor.go

// Idiomatic: The consumer defines exactly what it needs.
// The concrete 'UserStore' doesn't even need to know this interface exists.
type UserFetcher interface {
    FetchUser(id string) (*User, error)
}

type Processor struct {
    fetcher UserFetcher
}
```

#### 2.1.4 接受接口，返回结构体

要求输入参数尽可能使用最小的接口（例如，使用 `io.Reader` 而不是 `*os.File`），
但返回具体结构体，以便调用者无需使用类型断言即可访问特定字段或方法。

当进行依赖注入时，通过接口注入（依赖接口），而非具体类。

#### 2.1.5 不要使用 DI 框架

DI 框架让程序丧失了自己创建和组合对象的能力，完全依赖外部环境的组装。

注册式本身没有问题，问题是将配置、组装完全泄露到实现层，
这迫使你必须阅读所有的代码，才能知道真正的依赖关系，
因为你根本无法知道到底谁在负责某个依赖的配置和组装。

应该尽可能在应用程序入口点附近手动组装对象依赖。

> **Dependency Injection in .NET** -- Mark Seemann
>
> Create your object graph as close to the application's entry point as possible.

### 2.2 配置与结构体设计

#### 2.2.1 使用函数选项实现复杂初始化

当结构体有许多可选配置参数时，避免使用庞大的构造函数。请使用函数选项模式。

```go
type Server struct {
    addr    string
    timeout time.Duration
}

type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) { s.timeout = d }
}

func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second, // Sane default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

### 2.3 开发和工程化

#### 2.3.1 反兼容性

在大多数情况下，我们不需要兼容，重构、重写代码就是为了更优雅的实现，
兼容性代码只会增加程序复杂度，让程序变得更加难以维护。

除非用户特别说明需要兼容，否则默认不应该编写兼容性代码，不要考虑兼容性方案，不要替用户考虑迁移复杂度。

如果一定需要兼容，应该使用专门的兼容层，防止腐蚀项目代码。

#### 2.3.2 不要将二进制编译到项目根目录

不要将二进制编译到项目根目录，而是编译到 dist/builds/temp 等不受 git 追踪的目录下。

#### 2.3.3 Makefile 使用

优先使用 Makefile 而不要自己构造编译/测试命令

#### 2.3.4 格式化

使用 goimports 格式化代码，而不是使用 gofmt（优先使用 make fmt 进行代码格式化）

## 3. 基础框架

见 [LAYOUT.md](./references/LAYOUT.md)

### 3.1 工程化

Makefile 编写规范见 [MAKEFILE.md](./references/MAKEFILE.md)

## 4. 基础库

基础库使用见 [KITS.md](./references/KITS.md)

### 4.1 路由库

涉及 HTTP API 时优先使用标准库，如果没有特殊需求，引入第三方路由只考虑使用与标准库 100% 兼容的库，
详细选型建议见 [ROUTER.md](./references/ROUTER.md)

## 5. 代码风格

见 [STYLE.md](./references/STYLE.md)

### 5.1 错误处理

见 [ERRORS.md](./references/ERRORS.md)

### 5.2 数据库操作

见 [DATABASE.md](./references/DATABASE.md)

## 6. 参考资料

- [Which Go Router Should I Use? - Alex Edwards](https://www.alexedwards.net/blog/which-go-router-should-i-use)
- [go-chi/chi GitHub](https://github.com/go-chi/chi)
- [gorilla/mux GitHub](https://github.com/gorilla/mux)
- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Go 官方模块布局指南](https://go.dev/doc/modules/layout)
- [DI Containers are Code Polluters - Yegor Bugayenko](https://www.yegor256.com/2014/10/03/di-containers-are-evil.html)
- [GORM 官方文档](https://gorm.io/docs/)
- [GORM gen 文档](https://github.com/go-gorm/gen)
- [ykit 源码](https://gitcode.yealink.com/server/devops/yops/ykit)
