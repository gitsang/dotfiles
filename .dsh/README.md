# dsh configs

本目录是 DeepSeek Harness 的数据根（`$DSH_HOME`，默认 `~/.dsh`）。
配置由 [dotfiles](https://github.com/gitsang/dotfiles)（`$HOME` 工作树 git 仓库）
追踪；插件全部从 npm / GitHub Release 安装，不依赖本机源码目录。

## 新机器安装

```bash
# 1. dotfiles（见 ~/README.md）
# 2. 运行时
npm i -g @deepseek-ai/dsh pnpm

# 3. 密钥（不进 git，从密码管理器/旧机器安全拷贝 ~/.dsh/.credentials.yaml）
#    refs: SAIGW_API_KEY / YEALINK_API_KEY；records 段首次启动自动生成

# 4. 一条命令激活全部 profile 插件
~/.dsh/bootstrap.sh
#    常驻服务：DSH_TRUSTED_HOST=<域名> DSH_PROXY=<如需> ~/.dsh/bootstrap.sh --daemon
```

> dsh 对未安装的插件 bundle 是**静默跳过**，clone + checkout 只恢复配置清单，
> 必须跑 bootstrap（或逐 profile `pnpm install`）插件才生效。
> systemd unit 含本机 node 路径，是生成物，由 `--daemon` 每台机器重新生成。

启动后：web `http://127.0.0.1:3080`，gateway `http://127.0.0.1:3081`
（签入 URL 见 `journalctl --user -u dsh-daemon`，token 持久化后重启不变）。

## 数据布局

划分轴是**作用域**（跟机器走 vs 跟 profile 走），不是"配置 vs 运行时"：

| 位置 | 内容 | git |
|---|---|---|
| `cordis.patch.yml` | home 层补丁：模型网关，所有 profile 继承 | ✅ |
| `profiles/<name>/` | profile 清单、锁文件、cordis.patch.yml | ✅ |
| `profiles/<name>/gateway/settings.json` | gateway 的 pwa/hostMode 开关 | ✅ |
| `profiles/<name>/auto-review-enhance/settings.json` | Auto review 选择 | ✅ |
| `gateway/token.txt` | gateway 持久签入 token（凭证，0600） | ❌ |
| `bridge/`、`daemon/` | 插件私有状态（跨 profile） | ❌ |
| `sessions/` `storages/` `attachments/` `cache/` `logs/` `workspaces/` | 官方共享数据/缓存 | ❌ |
| `.credentials.yaml` | API 密钥 | ❌ 绝不进 git |

各插件对旧位置的文件会在首次读取时自动搬移到上表位置（现有文件优先）。

## 插件更新流程

源码仓库（`~/src/github.com/gitsang/dsh-*`）改完 → bump version →
`npm publish --access public --userconfig ~/.npmrc-publish`
（qol 用 `pnpm -r publish`）→ 各 profile `pnpm update <包>`。

pnpm 对 qol 子包的传递依赖更新可能不动锁文件，此时用
`pnpm add <包>@<版本>` 提为直接依赖钉住。

发布用 npm token 在 `~/.npmrc-publish`（0600，不进 git）。