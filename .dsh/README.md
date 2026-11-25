# dsh 配置（随 dotfiles 携带）

本目录是 DeepSeek Harness 的数据根（`$DSH_HOME`，默认 `~/.dsh`）。
配置部分由 [dotfiles](https://github.com/gitsang/dotfiles) bare 仓库追踪，
插件全部从 npm / GitHub Release 安装，不依赖任何本机源码目录。

## 新机器安装

### 1. 恢复 dotfiles

```bash
git clone --bare git@github.com:gitsang/dotfiles.git "$HOME/.dotfiles"
git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" config --local status.showUntrackedFiles no
git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" checkout
```

（`dotfiles` 命令别名见 `~/README.md`。）

### 2. 安装运行时

```bash
# node 用 fnm 管理（unit 里的 node 路径按机器生成，无需手改）
npm i -g @deepseek-ai/dsh pnpm
```

### 3. 放置密钥（不进 git，需手动）

从密码管理器或旧机器安全拷贝 `~/.dsh/.credentials.yaml`：

```yaml
version: 1
refs:
  SAIGW_API_KEY: <your key>
  YEALINK_API_KEY: <your key>
```

`records` 段（browser-session 签名密钥）由 dsh 首次启动自动生成，不用拷。

### 4. 一条命令激活

```bash
~/.dsh/bootstrap.sh
```

脚本会：检查 node/pnpm/dsh → 检查密钥文件 → 给 `profiles/` 下每个
profile 执行 `pnpm install` → `--dump-config` 校验 web/daemon/gateway。

> 为什么必须跑这一步：dsh 启动时对未安装的插件 bundle 是**静默跳过**
> （stderr 打一行 `skipping profile bundle ...`），UI 能起但插件全部消失。
> clone + checkout 只恢复配置清单，插件本体要靠 pnpm install。

### 5. 常驻服务（可选）

```bash
DSH_TRUSTED_HOST=<本机反代域名> DSH_PROXY=<如需出站代理> ~/.dsh/bootstrap.sh --daemon
# 等价于：dsh --profile daemon install --trusted-host <域名> --proxy <URL>
```

systemd unit（`~/.config/systemd/user/dsh-daemon.service`）是生成物，
含本机 node 绝对路径，**不进 git**，每台机器重新生成。
建议 `loginctl enable-linger` 让服务在注销后存活。

启动后：web UI `http://127.0.0.1:3080`，gateway `http://127.0.0.1:3081`
（签入 URL 由 gateway 在启动日志打印：`journalctl --user -u dsh-daemon`）。

## 目录布局约定

划分轴是**作用域**（跟机器走 vs 跟 profile 走），不是"配置 vs 运行时"：

| 位置 | 内容 | git |
|---|---|---|
| `cordis.patch.yml` | home 层补丁：模型网关，所有 profile 继承 | ✅ |
| `profiles/<name>/` | profile 配置：package.json、锁文件、cordis.patch.yml | ✅ |
| `profiles/web/.auto-review-enhance.json` | Auto review 的 profile 级选择 | ✅ |
| `profiles/gateway/settings.json` | gateway 的 pwa/hostMode 开关 | 由插件 .gitignore 管理 |
| `bridge/`、`daemon/` | 插件私有状态（跨 profile），首读自动从旧平铺文件迁移 | ❌ |
| `sessions/` `storages/` `attachments/` `cache/` `logs/` `workspaces/` | 官方共享数据/缓存 | ❌ |
| `.credentials.yaml` | API 密钥 | ❌ 绝不进 git |
| `profiles/gateway/{auth,session}.json`、`token.txt` | 凭证 | ❌ |

## 插件更新流程

源码仓库（`~/src/github.com/gitsang/dsh-*`）改完 → bump version →
`npm publish --access public --userconfig ~/.npmrc-publish`
（qol 用 `pnpm -r publish`）→ 各 profile `pnpm update <包>`。

pnpm 对 qol 子包的传递依赖更新可能不动锁文件，此时用
`pnpm add <包>@<版本>` 提为直接依赖钉住。

发布用的 npm token 在 `~/.npmrc-publish`（0600，不进 git）。