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

### 1.2 Install `dotfiles` command

Add function above into `~/.shrc`、`~/.bashrc` or `~/.zshrc`:

```bash
dotfiles() {
  git --git-dir="$HOME/.dotfiles" --work-tree="$HOME" "$@"
}

# let plain `git` work inside tracked directories, see 1.2.1
if [ -d "$HOME/.dotfiles" ] && [ ! -e "$HOME/.git" ]; then
  printf 'gitdir: %s/.dotfiles\n' "$HOME" > "$HOME/.git"
  git --git-dir="$HOME/.dotfiles" config core.bare false
  git --git-dir="$HOME/.dotfiles" config core.worktree "$HOME"
fi
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

## 3. Plain `git` inside tracked directories

Plain `git status`, `git diff`, `git add`, `git commit`, `git push` all work without the `dotfiles` prefix.

| Command                         | Purpose                                                                                                                                                                                                  |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `printf 'gitdir: ...' > ~/.git` | Put a **text pointer file** in `~` (git's native gitfile mechanism, same as submodules and `git worktree`). Walking up from a subdirectory, git finds it and learns the repo data lives in `~/.dotfiles` |
| `config core.bare false`        | Lift the "bare repo has no work tree" restriction, otherwise `status`/`add`/`checkout` fail with `fatal: this operation must be run in a work tree`. The `.dotfiles` layout is unchanged                 |
| `config core.worktree "$HOME"`  | Declare `~` as the work tree, so paths resolve relative to `$HOME`. Must be paired with `core.bare false`, otherwise git reports `core.bare and core.worktree do not make sense`                         |

Caveats:

- Any directory under `~` **without its own `.git`** (`~/Downloads`, `~/tmp`, ...) also
  resolves to the dotfiles repo, so `git add .` there would stage unrelated files.
  Nested real repositories are unaffected: git walks up from the cwd and uses whichever
  `.git` it hits first.
- `status.showUntrackedFiles=no` still applies, so plain `git status` does not flood the
  screen with untracked files from the whole home directory.
- These three lines **cannot be committed**: git rejects any path named `.git` at the
  index level, and `core.bare` / `core.worktree` live in `.dotfiles/config`, which `clone`
  regenerates locally. That is why the shell snippet recreates them on each machine,
  skipping the work when `~/.git` already exists.
- Some tools scan for `$HOME/.git` and treat the entire home directory as a git project
  (editors, prompt themes, ...), which may be slow or behave oddly.
- In `$HOME`, be careful with `git clean`.
- To undo: `rm ~/.git`, then
  `git --git-dir="$HOME/.dotfiles" config core.bare true` and
  `git --git-dir="$HOME/.dotfiles" config --unset core.worktree`, and remove the `if`
  block from your shell rc. Deleting only `~/.git` makes the next shell recreate it.
