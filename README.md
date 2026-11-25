<!-- MarkdownTitleNumber auto -->

# Dotfiles

Manage dotfiles via Git bare repository

## 1. Install in New Machine

### 1.1 Clone bare repository

```bash
git clone --bare git@github.com:gitsang/dotfiles.git "$HOME/.dotfiles"

# not showing untracked files
git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" config --local status.showUntrackedFiles no
```

### 1.2 Using `dotfiles` command

#### 1.2.1 Install

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

#### 1.2.2 Restore dotfiles

```bash
dotfiles reset HEAD .
dotfiles checkout .
```

### 1.3 (Optional) Using plain `git` inside tracked directories

#### 1.3.1 Configure

```sh
if [ -d "$HOME/.dotfiles" ] && [ ! -e "$HOME/.git" ]; then
  printf 'gitdir: %s/.dotfiles\n' "$HOME" > "$HOME/.git"
  git --git-dir="$HOME/.dotfiles" config core.bare false
  git --git-dir="$HOME/.dotfiles" config core.worktree "$HOME"
fi
```

> [!WARNING]
>
> - In `$HOME`, be careful with `git clean`.
> - Any directory under `~` **without its own `.git`** (`~/Downloads`, `~/tmp`, ...) also
>   resolves to the dotfiles repo, so `git add .` there would stage unrelated files.
>   Nested real repositories are unaffected: git walks up from the cwd and uses whichever
>   `.git` it hits first.
> - `status.showUntrackedFiles=no` still applies, so plain `git status` does not flood the
>   screen with untracked files from the whole home directory.
> - Some tools scan for `$HOME/.git` and treat the entire home directory as a git project
>   (editors, prompt themes, ...), which may be slow or behave oddly.
> - To undo: `rm ~/.git`, then
>   `git --git-dir="$HOME/.dotfiles" config core.bare true` and
>   `git --git-dir="$HOME/.dotfiles" config --unset core.worktree`, and remove the `if`
>   block from your shell rc. Deleting only `~/.git` makes the next shell recreate it.

#### 1.3.2 Restore dotfiles

```bash
git reset HEAD .
git checkout .
```
