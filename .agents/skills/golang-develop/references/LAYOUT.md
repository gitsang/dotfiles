<!-- MarkdownTitleNumber auto -->

# Golang 基础框架

## 1. 目录结构

整体目录结构应遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout) 社区约定。

> [!IMPORTANT]
>
> 这**不是官方标准**，而是社区总结的最佳实践。简单项目无需照搬全部结构。
>
> 比如小项目可以从单文件 `main.go` 开始，逐步增加目录

请到 [layout](./layout/) 目录查看基础的结构示例和目录说明。

## 2. 分层规则

默认按下面几层组织：

```text
cmd -> internal/<app>/app -> handler -> service -> repository
```

- `cmd`：进程入口，只做参数、配置、日志和启动退出，不组装业务依赖图。
- `app`：应用装配和生命周期，创建 handler、service、repository 以及数据库、缓存、MQ 等资源。
- `handler`：协议适配层，放 `handler.go`、`register.go`、`{name}.handler.go`。只做参数绑定、校验、调用 service、响应和错误码映射。
- `service`：业务用例层，放 `service.go`、`{name}.svc.go`。负责业务流程、事务边界、幂等、权限和一致性，不依赖 HTTP/gRPC 类型。
- `repository`：存储适配层，负责数据库、缓存等持久化访问，对 service 屏蔽存储细节。

简单项目不需要一次性建空目录；从一个用例开始，按职责自然拆出来。

## 3. Makefile 规范

Makefile 编写规范请参阅 [MAKEFILE.md](./MAKEFILE.md)。
