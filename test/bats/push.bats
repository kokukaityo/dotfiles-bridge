setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "push: auto カテゴリの変更を commit → push" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"

  echo "new content" > "$DOTFILES_DIR/editor/settings.json"

  run dotfile push
  assert_success
  assert_output --partial "Pushed:"
}

@test "push: 変更なしならスキップ" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"

  run dotfile push
  assert_success
  assert_output --partial "変更はありません"
}

@test "push: デフォルトブランチ以外ではスキップ" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"

  git -C "$DOTFILES_DIR" checkout -b feature/test --quiet

  run dotfile push
  assert_success
  assert_output --partial "自動pushをスキップ"
}
