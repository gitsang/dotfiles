<!-- MarkdownTitleNumber auto -->

# Dotfiles

Manage dotfiles with a Git repository whose work tree is $HOME

> [!WARNING]
>
> - In `$HOME`, be careful with `git clean`.

## 1. Install in New Machine

```bash
git clone --bare git@github.com:gitsang/dotfiles.git "$HOME/.git"
git --git-dir="$HOME/.git" config core.bare false
git --git-dir="$HOME/.git" config core.worktree "$HOME"
git -C "$HOME" reset HEAD .
git -C "$HOME" checkout .
```

> [!NOTE]
>
> To track a new top-level path, add `!/xxx` to `~/.gitignore`
