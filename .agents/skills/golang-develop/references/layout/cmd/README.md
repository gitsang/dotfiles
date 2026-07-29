# `cmd/`

## 目录结构

`cmd/` 是可执行程序的入口。

除非你的项目足够小，你可以将 `main.go` 放在根目录，否则，你应该将 `main.go` 放在 `cmd/` 目录。

最极端情况（当你的项目拥有多个应用程序，每个应用程序包含多个可执行程序）目录结构应该如下：

```
└── cmd
    ├── app1
    │   ├── cmd1
    │   │   └── main.go
    │   └── cmd2
    │       └── main.go
    └── app1
        ├── cmd1
        │   └── main.go
        └── cmd2
            └── main.go
```

但不必总是按照这个目录结构，如果你的项目只包含一个应用和一个可执行程序，通常可以直接将 `main.go` 放在 `cmd/` 目录中。

```
└── cmd
    └── main.go
```

而比较常见的情况是你的项目包含多个可执行程序（如果只有一个可执行程序，也推荐按照以下模式），比如：

```
└── cmd
    ├── petstore
    │   └── main.go
    └── petools
        └── main.go
```

不要在这个目录中放置太多代码。通常有一个小的 main 函数，进行简单的进程级别的组装，然后调用 `internal/` 或 `pkg/` 中的代码执行。

## 应该写什么

`cmd/` 只做 process-level wiring，不做 application-level wiring 和业务逻辑。

使用 `github.com/spf13/cobra` 创建命令，
使用 `gitcode.yealink.com/server/devops/yops/ykit/configer` 自动绑定命令行参数、环境变量、配置文件的配置。

请按照 [petstore/](petstore/) 示例进行代码编写。

目录内通常包含 `main.go`，`root.go` 和 `<subcommand>.go` 文件。

### `main.go`

`main.go` 应该非常简单，只有一个 main 函数，只调用 `rootCmd.Execute()`。

### `root.go`

它应该包含：

1. `rootCmd`
   - 定义命令名、说明、默认 `Run`。
   - 命令名要和目录名一致，例如 `cmd/petstore` 的 `Use: "petstore"`。

2. 全局 flags
   - 至少保留 `--config-file` / `-c`。
   - 多应用项目的默认配置路径应指向 `configs/<cmd_name>/config.yml`；单应用项目用 `configs/config.yml`。
   - 业务配置不要手写成一堆独立 flag；让 `internal/config.NewConfiger` 或 `internal/<cmd_name>/config.NewConfiger` 根据配置结构自动绑定 `--config.*`。

3. `init()`
   - 注册持久化 flags。
   - 创建 configer：`cfger = config.NewConfiger(rootCmd)`。
   - 注册子命令，例如 `version`。

4. `Run()`
   - 创建根 context。
   - 加载配置文件。
   - 初始化日志 handler 和 `slog.Logger`。
   - 调用 `app.New(ctx, cfg, logger)`。
   - 注册 graceful shutdown。
   - 调用 `app.Start()`。

### `version.go`

version 是常见的子命令，通常应该总是包含这个子命令作为版本展示，其中定义的全局变量将在编译时通过 LD_FLAGS 注入。

## 不应该写什么

- 不直接初始化数据库、cache、mq、HTTP routes。
- 不直接组装 handler、service、repository、worker 等应用依赖图。
- 不写业务逻辑。
- 不把配置解析散落在入口函数里。
- 不让 `cmd` 依赖 `internal/...` 之外的低层实现包，除非是日志、版本等进程启动必需品。
