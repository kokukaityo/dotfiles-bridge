setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "delete-category: auto カテゴリを削除して push" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  echo "data" > "$DOTFILES_DIR/editor/settings.json"
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add editor file" --quiet
  git -C "$DOTFILES_DIR" push origin main --quiet 2>/dev/null

  local before_count
  before_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)

  run dotfile delete-category editor
  assert_success
  assert_output --partial "カテゴリを削除しました"

  assert_file_not_exists "$DOTFILES_DIR/editor"

  run cat "$DOTFILES_DIR/sync.toml"
  refute_output --partial '"editor"'

  local after_count
  after_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)
  [[ "$((after_count - before_count))" -eq 1 ]] || fail "expected exactly 1 new commit"

  run git -C "$DOTFILES_DIR" log -1 --format=%s
  assert_output --partial "auto category editor"
}

@test "delete-category: manual カテゴリを削除して push" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"

  mkdir -p "$DOTFILES_DIR/docs"
  echo "manual content" > "$DOTFILES_DIR/docs/guide.md"
  cat > "$DOTFILES_DIR/sync.toml" <<TOML
default_branch = "main"
mode = "remote"
auto = ["editor"]
ignore = ["backup"]
TOML
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add docs" --quiet
  git -C "$DOTFILES_DIR" push origin main --quiet 2>/dev/null

  local before_count
  before_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)

  run dotfile delete-category docs
  assert_success
  assert_output --partial "カテゴリを削除しました"

  assert_file_not_exists "$DOTFILES_DIR/docs"

  local after_count
  after_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)
  [[ "$((after_count - before_count))" -eq 1 ]] || fail "expected exactly 1 new commit"

  run git -C "$DOTFILES_DIR" log -1 --format=%s
  assert_output --partial "manual category docs"
}

@test "delete-category: ignore カテゴリを削除して push" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  # .gitignore を生成して commit
  run dotfile gitignore
  assert_success
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add gitignore" --quiet
  git -C "$DOTFILES_DIR" push origin main --quiet 2>/dev/null

  # ignore カテゴリのディレクトリを作成（Git 追跡外）
  mkdir -p "$DOTFILES_DIR/backup"
  echo "backup data" > "$DOTFILES_DIR/backup/db.sql"

  local before_count
  before_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)

  run dotfile delete-category backup
  assert_success
  assert_output --partial "カテゴリを削除しました"

  assert_file_not_exists "$DOTFILES_DIR/backup"

  run cat "$DOTFILES_DIR/sync.toml"
  refute_output --partial '"backup"'

  run cat "$DOTFILES_DIR/.gitignore"
  refute_output --partial 'backup/'

  local after_count
  after_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)
  [[ "$((after_count - before_count))" -eq 1 ]] || fail "expected exactly 1 new commit"

  run git -C "$DOTFILES_DIR" log -1 --format=%s
  assert_output --partial "ignore category backup"
}

@test "delete-category: 未登録カテゴリのディレクトリを削除して push" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  mkdir -p "$DOTFILES_DIR/orphan"
  echo "orphan data" > "$DOTFILES_DIR/orphan/file.txt"
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add orphan" --quiet
  git -C "$DOTFILES_DIR" push origin main --quiet 2>/dev/null

  local before_count
  before_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)

  run dotfile delete-category orphan
  assert_success
  assert_output --partial "カテゴリを削除しました"

  assert_file_not_exists "$DOTFILES_DIR/orphan"

  local after_count
  after_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)
  [[ "$((after_count - before_count))" -eq 1 ]] || fail "expected exactly 1 new commit"

  run git -C "$DOTFILES_DIR" log -1 --format=%s
  assert_output "delete: manual category orphan"
}

@test "delete-category: 存在しないカテゴリはエラー" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  run dotfile delete-category nonexistent
  assert_failure
  assert_output --partial "カテゴリが見つかりません"
}

@test "delete-category: 不正なカテゴリ名はエラー" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  run dotfile delete-category "../escape"
  assert_failure
  assert_output --partial "不正なカテゴリ名"
}
