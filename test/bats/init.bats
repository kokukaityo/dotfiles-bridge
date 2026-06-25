setup() {
  load helper/setup
  _common_setup
  TARGET="$TEST_TEMP_DIR/my-dotfiles"
}

teardown() {
  _common_teardown
}

@test "init: 指定パスにデータリポジトリを新規作成" {
  can_symlink || skip "init runs link internally; symlink not available"

  run dotfiles init "$TARGET"
  assert_success
  assert_output --partial "データリポジトリを作成"
  assert_output --partial "作成が完了"

  assert_file_exists "$TARGET/sync.toml"
  assert_file_exists "$TARGET/.infra-version"
  assert_file_exists "$TARGET/.gitignore"

  run cat "$TARGET/.infra-version"
  assert_output --partial "1.0.0"

  assert_file_exists "$TARGET/ai-agent/link.toml"
  assert_file_exists "$TARGET/editor/link.toml"
  assert_file_exists "$TARGET/shell/link.toml"

  run git -C "$TARGET" log --oneline
  assert_success
  assert_output --partial "initial dotfiles setup"

  run git -C "$TARGET" status --porcelain
  assert_success
  assert_output ""
}

@test "init: hooks と git 設定が適用される" {
  can_symlink || skip "init runs link internally; symlink not available"

  dotfiles init "$TARGET"

  assert_file_exists "$TARGET/.dotfiles-hook/pre-push"
  assert_file_exists "$TARGET/.dotfiles-hook/post-merge"

  run git -C "$TARGET" config core.hooksPath
  assert_output ".dotfiles-hook"

  assert_file_contains "$TARGET/.gitattributes" '* -text'
}

@test "init: 既存パスへの init はエラー" {
  mkdir -p "$TARGET"
  run dotfiles init "$TARGET"
  assert_failure
  assert_output --partial "既にパスが存在します"
}
