# Makefile 编写规范

本文档描述 Go 项目 Makefile 的编写规范。

## Makefile 结构规范

Makefile 应按以下顺序组织：

```makefile
SHELL := /usr/bin/env bash

# 版本信息
# Go 环境
# 项目特定变量
# 工具安装
# 目标分类（使用 ##@ 注释）
```

### 1. 版本信息

```makefile
VERSION    ?= $(shell git describe --tags --always --match "v*" --dirty 2>/dev/null || git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u --iso-8601=seconds)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
```

- `VERSION`：使用 `git describe` 自动生成，支持语义化版本
- `BUILD_DATE`：UTC 时间戳，ISO 8601 格式
- `GIT_COMMIT`：当前提交哈希

### 2. Go 环境变量

```makefile
GO          := go
GO_VERSION  := $(shell $(GO) version | awk '{print $$3}')
GOOS        := $(if $(GOOS),$(GOOS),$(shell $(GO) env GOOS))
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell $(GO) env GOARCH))
GOPATH      := $(shell $(GO) env GOPATH)
```

- 使用 `:=` 立即求值
- `GOOS`/`GOARCH` 支持外部覆盖

### 3. 工具安装模式

```makefile
BUILD_DIR ?= $(PWD)/.builds
BUILD_BIN ?= $(BUILD_DIR)/bin
$(BUILD_BIN):
	mkdir -p $(BUILD_BIN)
export PATH := $(BUILD_BIN):$(PATH)

# 工具名
TOOL         = $(BUILD_BIN)/tool-name
TOOL_VERSION = v1.0.0
$(TOOL): $(BUILD_BIN)
ifeq ($(wildcard $(TOOL)),)
	GOBIN=$(BUILD_BIN) CGO_ENABLED=0 $(GO) install example.com/tool@$(TOOL_VERSION)
endif
```

**规范：**
- 工具安装到 `.builds/bin` 目录
- 使用 `GOBIN` 指定安装路径
- 使用 `CGO_ENABLED=0` 静态编译
- 使用 `wildcard` 检查是否已安装
- 版本号明确指定，避免使用 `latest`

### 4. 目标分类

使用 `##@` 注释创建帮助分组：

```makefile
#------------------------------------------------------------------------------#
##@ Prerequisites
#------------------------------------------------------------------------------#

#------------------------------------------------------------------------------#
##@ Generate
#------------------------------------------------------------------------------#

#------------------------------------------------------------------------------#
##@ Develop
#------------------------------------------------------------------------------#

#------------------------------------------------------------------------------#
##@ Build
#------------------------------------------------------------------------------#

#------------------------------------------------------------------------------#
##@ Publish
#------------------------------------------------------------------------------#

#------------------------------------------------------------------------------#
##@ Clean
#------------------------------------------------------------------------------#

#------------------------------------------------------------------------------#
##@ Help
#------------------------------------------------------------------------------#
```

### 5. 目标注释

- 使用 `##` 注释目标描述
- 格式：`## description`
- 示例：`## build app-name binary`

```makefile
.PHONY: build-app-name
## build app-name binary
build-app-name: $(APP_NAME_DIST)
	...
```

### 6. 版本注入

通过 LDFLAGS 注入版本信息到 Go 二进制文件：

```makefile
# 每个应用定义自己的 LDFLAGS
APP_NAME_LDFLAGS += -s -w
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.Version=$(VERSION)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.BuildDate=$(BUILD_DATE)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GitCommit=$(GIT_COMMIT)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GoVersion=$(GO_VERSION)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GOOS=$(GOOS)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GOARCH=$(GOARCH)"
```

- `-s -w`：去除符号表和调试信息
- 使用完整模块路径，包括应用子目录

### 7. 目录结构

```makefile
# dist
OS        ?= linux
ARCH      ?= amd64
DIST_ROOT ?= dist

# 应用特定变量（每个应用定义自己的变量）
# app-name
APP_NAME ?= app-name
# app-name:dist
APP_NAME_DIST      ?= $(DIST_ROOT)/$(APP_NAME)/$(OS)/$(ARCH)
APP_NAME_DIST_BIN  ?= $(APP_NAME_DIST)/bin
APP_NAME_DIST_CONFIG ?= $(APP_NAME_DIST)/configs
```

- 按 `dist/<app>/<os>/<arch>/` 组织构建产物
- 每个应用有独立的变量定义
- 使用 `rsync` 同步配置文件

### 8. 帮助生成

```makefile
HELP_TAB_LEN := 36
## display help
help:
	@awk 'BEGIN \
	{ \
		FS = ":.*##"; \
		printf "\nUsage:\n  make \033[36m<target>\033[0m\n" \
	} \
	/^[0-9a-zA-Z\_\-\/]+:/ \
	{ \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "  \033[36m%-$(HELP_TAB_LEN)s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} { lastLine = $$0 } \
	/^##@/ \
	{ \
		printf "\n\033[1m%s\033[0m\n", substr($$0, 5) \
	} ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
```

- 默认目标为 `help`
- 自动从注释生成帮助信息

## 多应用目标约定

当项目包含多个应用时，应遵循以下目标命名约定：

### 应用特定目标

所有应用特定目标都带有应用名后缀：

- `build-<app>`：构建单个应用的二进制文件
- `build-artifacts-<app>`：构建单个应用的所有本地制品
- `build-container-<app>`：构建单个应用的容器镜像
- `run-<app>`：运行单个应用
- `dev-<app>`：开发模式运行单个应用（CGO_ENABLED=1）
- `debug-<app>`：调试运行单个应用
- `publish-<app>`：发布单个应用的容器镜像
- `publish-release-<app>`：发布单个应用到发布仓库

### 聚合目标

不带后缀的目标用于操作所有应用：

- `build-artifacts`：构建所有应用的本地制品
- `build-container`：构建所有应用的容器镜像
- `build`：构建所有（制品 + 容器镜像）
- `publish`：发布所有应用的容器镜像
- `publish-release`：发布所有应用到发布仓库

### 应用变量定义

每个应用应定义以下变量：

```makefile
# app-name
APP_NAME ?= app-name
# app-name:go
APP_NAME_LDFLAGS += -s -w
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.Version=$(VERSION)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.BuildDate=$(BUILD_DATE)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GitCommit=$(GIT_COMMIT)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GoVersion=$(GO_VERSION)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GOOS=$(GOOS)"
APP_NAME_LDFLAGS += -X "module/path/cmd/app-name.GOARCH=$(GOARCH)"
# app-name:dist
APP_NAME_DIST      ?= $(DIST_ROOT)/$(APP_NAME)/$(OS)/$(ARCH)
APP_NAME_DIST_BIN  ?= $(APP_NAME_DIST)/bin
APP_NAME_DIST_CONFIG ?= $(APP_NAME_DIST)/configs
```

### 添加新应用

要添加新应用，请按照以下步骤操作：

#### 1. 定义应用变量

在 Makefile 的变量定义区域添加新应用的变量：

```makefile
# myapp
MYAPP ?= myapp
# myapp:go
MYAPP_LDFLAGS += -s -w
MYAPP_LDFLAGS += -X "golang-develop-example/cmd/myapp.Version=$(VERSION)"
MYAPP_LDFLAGS += -X "golang-develop-example/cmd/myapp.BuildDate=$(BUILD_DATE)"
MYAPP_LDFLAGS += -X "golang-develop-example/cmd/myapp.GitCommit=$(GIT_COMMIT)"
MYAPP_LDFLAGS += -X "golang-develop-example/cmd/myapp.GoVersion=$(GO_VERSION)"
MYAPP_LDFLAGS += -X "golang-develop-example/cmd/myapp.GOOS=$(GOOS)"
MYAPP_LDFLAGS += -X "golang-develop-example/cmd/myapp.GOARCH=$(GOARCH)"
# myapp:dist
MYAPP_DIST      ?= $(DIST_ROOT)/$(MYAPP)/$(OS)/$(ARCH)
MYAPP_DIST_BIN  ?= $(MYAPP_DIST)/bin
MYAPP_DIST_CONFIG ?= $(MYAPP_DIST)/configs
```

#### 2. 添加运行目标

在 Develop 区域添加运行目标：

```makefile
.PHONY: run-myapp
## run myapp app
run-myapp:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
		$(GO) run \
			-trimpath -v -ldflags '$(MYAPP_LDFLAGS)' \
			./cmd/$(MYAPP) \
			$(filter-out $@, $(MAKECMDGOALS))

.PHONY: debug-myapp
## debug run myapp app
debug-myapp: $(DELVE)
	$(DELVE) debug \
		./cmd/$(MYAPP) \
		--log \
		--output /tmp/__debug_$(MYAPP) \
		-- $(filter-out $@, $(MAKECMDGOALS))
```

#### 3. 添加构建目标

在 Build 区域添加构建目标：

```makefile
# myapp dist directories
$(MYAPP_DIST_BIN):
	@rm -fr $@ && mkdir -p $@

$(MYAPP_DIST_CONFIG):
	@rm -fr $@ && mkdir -p $@

$(MYAPP_DIST): $(MYAPP_DIST_BIN) $(MYAPP_DIST_CONFIG)
	@mkdir -p $@

.PHONY: build-artifacts-myapp
## build myapp all local artifacts
build-artifacts-myapp: $(MYAPP_DIST)
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) \
		$(GO) build \
			-trimpath -v -ldflags '$(MYAPP_LDFLAGS)' \
			-o $(MYAPP_DIST_BIN)/$(MYAPP) \
			./cmd/$(MYAPP)
	rsync -a --delete configs/config.example.yml $(MYAPP_DIST_CONFIG)/config.yml 2>/dev/null || true

.PHONY: build-container-myapp
## build myapp container image
build-container-myapp:
	$(CONTAINER_CLI) build \
		--build-arg DOCKER_PACKAGE_PATH="$(MYAPP_DIST)" \
		-t "$(CONTAINER_REPO)/$(MYAPP):$(subst +,-,$(VERSION))" \
		-f "Containerfile.$(MYAPP)" .
```

#### 4. 添加发布目标

在 Publish 区域添加发布目标：

```makefile
.PHONY: publish-myapp
## publish myapp container
publish-myapp:
	$(CONTAINER_CLI) push $(CONTAINER_REPO)/$(MYAPP):$(subst +,-,$(VERSION))

.PHONY: publish-release-myapp
## publish myapp container to release repo
publish-release-myapp:
	$(CONTAINER_CLI) tag $(CONTAINER_REPO)/$(MYAPP):$(subst +,-,$(VERSION)) $(CONTAINER_REPO_RELEASE)/$(MYAPP):$(subst +,-,$(VERSION))
	$(CONTAINER_CLI) push $(CONTAINER_REPO_RELEASE)/$(MYAPP):$(subst +,-,$(VERSION))
```

#### 5. 更新聚合目标

将新应用添加到聚合目标的依赖列表中：

```makefile
.PHONY: build-artifacts
## build all applications local artifacts
build-artifacts: build-artifacts-petstore build-artifacts-myapp

.PHONY: build-container
## build all container images
build-container: build-container-petstore build-container-myapp

.PHONY: publish
## publish all containers
publish: publish-petstore publish-myapp

.PHONY: publish-release
## publish all containers to release repo
publish-release: publish-release-petstore publish-release-myapp
```

## 最佳实践

### 变量命名
- 使用大写字母和下划线：`APP_NAME`
- 工具变量：`TOOL_NAME`、`TOOL_NAME_VERSION`
- 路径变量：`*_DIR`、`*_BIN`、`*_DIST`

### 依赖管理
- 使用 Make 规则表达依赖关系
- 工具依赖通过文件时间戳管理
- 使用 `.PHONY` 声明伪目标

### 命令执行
- 使用 `\` 续行，缩进使用 Tab
- 复杂命令使用 `\` 分行，每行一个操作
- 使用 `@` 抑制命令回显

### 错误处理
- 使用 `set -e` 确保脚本错误退出
- 检查命令返回值
- 提供有意义的错误信息

## 示例

参考 [layout/Makefile](./layout/Makefile) 了解完整实现。