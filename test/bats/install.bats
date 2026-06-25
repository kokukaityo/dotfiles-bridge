setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "install: clone 済みリポジトリに設定を適用" {
  create_data_repo
  run dotfiles install
  assert_success
  assert_output --partial "Setup complete."
}

@test "install: hooks と git 設定が適用される" {
  create_data_repo
  dotfiles install

  assert_file_exists "$DOTFILES_DIR/.dotfiles-hook/pre-push"
  assert_file_exists "$DOTFILES_DIR/.dotfiles-hook/post-merge"

  run git -C "$DOTFILES_DIR" config core.hooksPath
  assert_output ".dotfiles-hook"

  assert_file_contains "$DOTFILES_DIR/.gitattributes" '* -text'
  assert_file_contains "$DOTFILES_DIR/.gitignore" 'auto-generated from sync.toml'
}

@test "install: 冪等に実行できる" {
  create_data_repo
  dotfiles install
  run dotfiles install
  assert_success
}
