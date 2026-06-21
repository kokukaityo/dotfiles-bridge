PROJECT_ROOT="$(cd "$BATS_TEST_DIRNAME/../.." && pwd)"

load "${PROJECT_ROOT}/node_modules/bats-support/load.bash"
load "${PROJECT_ROOT}/node_modules/bats-assert/load.bash"

assert_file_exists() {
  [[ -e "$1" ]] || fail "expected file to exist: $1"
}

assert_file_not_exists() {
  [[ ! -e "$1" ]] || fail "expected file NOT to exist: $1"
}

assert_file_contains() {
  [[ -f "$1" ]] || fail "file does not exist: $1"
  grep -qF "$2" "$1" || fail "expected '$1' to contain '$2'"
}

_common_setup() {
  TEST_TEMP_DIR="$BATS_TEST_TMPDIR/case-$BATS_TEST_NUMBER"
  mkdir -p "$TEST_TEMP_DIR"

  export HOME="$TEST_TEMP_DIR/fakehome"
  mkdir -p "$HOME"

  # Windows: Go の os.UserHomeDir() は USERPROFILE を参照する
  if [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == MSYS* ]]; then
    export USERPROFILE="$(cygpath -w "$HOME")"
  fi

  export DOTFILES_DIR="$HOME/dotfiles"
  export GIT_CONFIG_NOSYSTEM=1
  export GIT_CONFIG_GLOBAL="$HOME/.gitconfig"
  export PATH="$PROJECT_ROOT/dist:$PATH"

  git config --global safe.bareRepository all
  git config --global init.defaultBranch main

  if [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == MSYS* ]]; then
    if [[ ! -f "$PROJECT_ROOT/dist/dotfile.exe" ]]; then
      fail "dist/dotfile.exe not found. Run 'make build' first."
    fi
  else
    if [[ ! -f "$PROJECT_ROOT/dist/dotfile" ]]; then
      fail "dist/dotfile not found. Run 'make build' first."
    fi
  fi
}

_common_teardown() {
  rm -rf "$TEST_TEMP_DIR"
}

get_os_key() {
  case "$(uname -s)" in
    MINGW*|MSYS*) echo "win32" ;;
    Darwin)       echo "darwin" ;;
    Linux)        echo "linux" ;;
    *)            echo "unknown" ;;
  esac
}

# Go の os.Symlink で NTFS symlink が作れるかテストする
can_symlink() {
  if [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == MSYS* ]]; then
    local test_target="$(cygpath -w "$TEST_TEMP_DIR/.symlink-target")"
    local test_link="$(cygpath -w "$TEST_TEMP_DIR/.symlink-test")"
    touch "$TEST_TEMP_DIR/.symlink-target"
    if cmd //c mklink "$test_link" "$test_target" >/dev/null 2>&1; then
      rm -f "$TEST_TEMP_DIR/.symlink-test" "$TEST_TEMP_DIR/.symlink-target"
      return 0
    fi
    rm -f "$TEST_TEMP_DIR/.symlink-target"
    return 1
  fi
  local test_target="$TEST_TEMP_DIR/.symlink-target"
  local test_link="$TEST_TEMP_DIR/.symlink-test"
  touch "$test_target"
  if ln -s "$test_target" "$test_link" 2>/dev/null; then
    rm -f "$test_link" "$test_target"
    return 0
  fi
  rm -f "$test_target"
  return 1
}

create_data_repo() {
  local repo="${1:-$DOTFILES_DIR}"
  local branch="${2:-main}"

  mkdir -p "$repo"
  git -C "$repo" init -b "$branch" --quiet
  git -C "$repo" config user.name "bats-test"
  git -C "$repo" config user.email "bats@test.invalid"

  echo "1.0.0" > "$repo/.infra-version"

  cat > "$repo/sync.toml" <<TOML
default_branch = "$branch"
auto = ["editor"]
manual = []
ignore = ["backup"]
TOML

  mkdir -p "$repo/editor"
  cat > "$repo/editor/link.toml" <<'TOML'
TOML

  git -C "$repo" add -A
  git -C "$repo" commit -m "initial" --quiet

  export DOTFILES_DIR="$repo"
}

create_bare_remote() {
  git init --bare "$1" --quiet
}

add_remote_to_repo() {
  local repo="$1" remote="$2" branch="$3"
  git -C "$repo" remote add origin "$remote"
  git -C "$repo" push -u origin "$branch" --quiet 2>/dev/null
  git --git-dir="$remote" symbolic-ref HEAD "refs/heads/$branch"
}
