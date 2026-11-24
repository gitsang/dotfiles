<!-- MarkdownTitleNumber auto -->

# Dotfiles

Manage dot files via Git bare repository

## 1. Install in New Machine

### 1.1 Clone bare repository

```bash
git clone --bare git@github.com:gitsang/dotfiles.git "$HOME/.dotfiles"

# not showing untracked files
git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" config --local status.showUntrackedFiles no
```

### 1.2 Install `dotfiles` command

Add function above into `~/.shrc`、`~/.bashrc` or `~/.zshrc`:

```bash
dotfiles() {
  git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" "$@"
}
```

reload shell configuration:

```bash
source ~/.shrc # or ~/.bashrc or ~/.zshrc
```

### 1.3 Restore dotfiles

```bash
dotfiles checkout
```

## 2. Use `dotfiles` command to replace `git`

```bash
dotfiles add .agents
dotfiles commit -m "Update agent skills"
dotfiles push
```

```bash
dotfiles pull
dotfiles checkout
```

```bash
dotfiles status
```

```bash
dotfiles remote -v
```

```bash
dotfiles diff --cached
```
