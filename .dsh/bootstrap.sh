#!/usr/bin/env bash
# Bootstrap this dsh home on a new machine after `dotfiles checkout`.
#
#   ~/.dsh/bootstrap.sh            install plugin deps for every profile
#   ~/.dsh/bootstrap.sh --daemon   also (re)install the dsh-daemon systemd unit
#
# Optional environment for --daemon (defaults match the dev1 machine):
#   DSH_TRUSTED_HOST   authority the web trust fence accepts (host[:port])
#   DSH_PROXY          outbound http(s) proxy URL, omit for none
#
# Not carried by dotfiles (by design): ~/.dsh/.credentials.yaml (API keys).
# The script checks it and tells you what to fill in; it never writes keys.
set -euo pipefail

DSH_HOME_DIR="${DSH_HOME:-$HOME/.dsh}"
PROFILES_DIR="$DSH_HOME_DIR/profiles"
WITH_DAEMON=0
[ "${1:-}" = "--daemon" ] && WITH_DAEMON=1

say()  { printf '\033[1;34m==>\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33mwarn:\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31merror:\033[0m %s\n' "$*" >&2; exit 1; }

# ---- 1. runtime prerequisites -------------------------------------------
command -v node >/dev/null || die "node not found; install node (fnm) first"
command -v pnpm >/dev/null || die "pnpm not found; run: npm i -g pnpm"
command -v dsh  >/dev/null || die "dsh not found; run: npm i -g @deepseek-ai/dsh"
say "node $(node --version), pnpm $(pnpm --version), dsh $(dsh --version 2>/dev/null | head -1)"

# ---- 2. credentials (never in git) ---------------------------------------
CREDS="$DSH_HOME_DIR/.credentials.yaml"
if [ ! -f "$CREDS" ]; then
  warn "missing $CREDS — model calls will fail until you create it:"
  cat >&2 <<'EOF'
    version: 1
    refs:
      SAIGW_API_KEY: <your key>
      YEALINK_API_KEY: <your key>
EOF
  warn "copy it from your password manager or the old machine, then re-run if needed"
else
  for key in SAIGW_API_KEY YEALINK_API_KEY; do
    grep -q "$key" "$CREDS" || warn "$key not found in $CREDS"
  done
  say "credentials present"
fi

# ---- 3. plugin dependencies per profile ----------------------------------
# dsh silently SKIPS bundles whose packages are not installed, so this step
# is what actually makes the plugins take effect.
[ -d "$PROFILES_DIR" ] || die "no profiles directory at $PROFILES_DIR; did dotfiles checkout run?"
installed=0
for dir in "$PROFILES_DIR"/*/; do
  [ -f "$dir/package.json" ] || continue
  name=$(basename "$dir")
  # profiles with no dependencies (acp/headless) resolve everything from the
  # dsh installation itself; installing them is still cheap and harmless.
  say "pnpm install: profile $name"
  (cd "$dir" && pnpm install --prefer-offline)
  installed=$((installed + 1))
done
say "installed dependencies for $installed profile(s)"

# ---- 4. composed-config sanity check -------------------------------------
for name in web daemon gateway; do
  [ -f "$PROFILES_DIR/$name/package.json" ] || continue
  if dsh --profile "$name" --dump-config >/dev/null 2>&1; then
    say "profile $name: composed config OK"
  else
    warn "profile $name: --dump-config failed; run 'dsh --profile $name --dump-config' to see why"
  fi
done

# ---- 5. optional systemd daemon ------------------------------------------
if [ "$WITH_DAEMON" = 1 ]; then
  args=()
  [ -n "${DSH_TRUSTED_HOST:-}" ] && args+=(--trusted-host "$DSH_TRUSTED_HOST")
  [ -n "${DSH_PROXY:-}" ] && args+=(--proxy "$DSH_PROXY")
  say "installing dsh-daemon systemd unit (restarts the web UI if running)"
  dsh --profile daemon install "${args[@]}"
  say "daemon installed; check: systemctl --user status dsh-daemon"
else
  say "done. Start the UI with 'dsh web', or re-run with --daemon for the systemd service"
fi