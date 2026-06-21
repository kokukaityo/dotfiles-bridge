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

  echo "data" > "$DOTFILES_DIR/editor/settings.json"
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add editor file" --quiet
  git -C "$DOTFILES_DIR" push origin main --quiet 2>/dev/null

  local before_count
  before_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)

  run dotfile delete-category editor
  assert_success
  assert_output --partial "カテゴリを削除してpush"

  assert_file_not_exists "$DOTFILES_DIR/editor"

  run cat "$DOTFILES_DIR/sync.toml"
  refute_output --partial '"editor"'

  local after_count
  after_count=$(git -C "$DOTFILES_DIR" rev-list --count HEAD)
  [[ "$((after_count - before_count))" -eq 1 ]] || fail "expected exactly 1 new commit"
}

@test "delete-category: auto 以外のカテゴリはエラー" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"

  run dotfile delete-category nonexistent
  assert_failure
  assert_output --partial "自動同期カテゴリではありません"
}

@test "delete-category: 不正なカテゴリ名はエラー" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"

  run dotfile delete-category "../escape"
  assert_failure
  assert_output --partial "不正なカテゴリ名"
}
