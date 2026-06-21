setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "setup: clone 済みリポジトリに設定を適用" {
  create_data_repo
  run dotfile setup
  assert_success
  assert_output --partial "Setup complete."
}

@test "setup: hooks と git 設定が適用される" {
  create_data_repo
  dotfile setup

  assert_file_exists "$DOTFILES_DIR/.dotfile-hook/pre-push"
  assert_file_exists "$DOTFILES_DIR/.dotfile-hook/post-merge"

  run git -C "$DOTFILES_DIR" config core.hooksPath
  assert_output ".dotfile-hook"

  assert_file_contains "$DOTFILES_DIR/.gitattributes" '* -text'
  assert_file_contains "$DOTFILES_DIR/.gitignore" 'auto-generated from sync.toml'
}

@test "setup: 冪等に実行できる" {
  create_data_repo
  dotfile setup
  run dotfile setup
  assert_success
}
