# Dotfiles

使用 Git bare repository 管理 home 目录中的配置和 Agent Skills。

## 当前内容

当前仓库追踪：

- `~/.agents/`
- Agent Skills 及其参考文件、脚本和锁文件

`node_modules` 等被 skill 自己的 `.gitignore` 排除的依赖不会提交。

## 在新机器上安装

### 1. 准备 GitHub SSH

确认当前机器可以访问 GitHub：

```bash
ssh -T git@github.com
```

如果还没有 SSH key：

```bash
ssh-keygen -t ed25519 -C "your-email@example.com"
cat ~/.ssh/id_ed25519.pub
```

将公钥添加到 GitHub 的 SSH keys 设置中。

### 2. 克隆 bare repository

```bash
git clone --bare git@github.com:gitsang/dotfiles.git "$HOME/.dotfiles"

git --git-dir="$HOME/.dotfiles" \\
  --work-tree="$HOME" \\
  config --local status.showUntrackedFiles no
```

### 3. 安装 `dotfiles` 命令

将下面的函数加入 `~/.shrc`、`~/.bashrc` 或 `~/.zshrc`：

```bash
dotfiles() {
  git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" "$@"
}
```

重新加载 shell 配置：

```bash
source ~/.shrc
```

如果使用 Bash 或 Zsh，请改为 source 对应的配置文件。

### 4. 恢复文件

先查看目标文件是否已经存在：

```bash
dotfiles status
```

如果目标机器上已经有 `.agents`，先备份：

```bash
mv ~/.agents ~/.agents.backup
```

然后恢复仓库中的文件：

```bash
dotfiles checkout
```

## 日常同步

在一台机器上修改或更新 skills 后：

```bash
dotfiles add .agents
dotfiles commit -m "Update agent skills"
dotfiles push
```

在其他机器上同步：

```bash
dotfiles pull
dotfiles checkout
```

查看当前状态：

```bash
dotfiles status
```

查看远程地址：

```bash
dotfiles remote -v
```

## 添加新的配置文件

只添加明确需要同步的文件，不要把整个 home 目录加入追踪：

```bash
dotfiles add ~/.vimrc
dotfiles add ~/.config/nvim/init.lua
dotfiles commit -m "Add editor configuration"
dotfiles push
```

路径会按照 `$HOME` 下的相对路径处理。

## 处理冲突

如果 `checkout` 因为目标文件已存在而失败，先备份本地文件，再重试：

```bash
mkdir -p ~/dotfiles-backup
mv ~/.agents ~/dotfiles-backup/
dotfiles checkout
```

如果远程和本地都发生了修改：

```bash
dotfiles pull
dotfiles status
dotfiles diff
```

解决文件冲突后重新提交：

```bash
dotfiles add <path>
dotfiles commit -m "Resolve dotfiles conflict"
dotfiles push
```

## 安全注意事项

不要提交以下内容：

- API keys、access tokens 和密码
- `.env` 文件
- SSH 私钥
- 云服务凭据
- 包含机器专属敏感信息的配置

提交前检查暂存内容：

```bash
dotfiles diff --cached
```
