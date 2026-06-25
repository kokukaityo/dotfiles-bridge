setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "status: コンフリクトなしなら無言で exit 0" {
  create_data_repo
  run dotfiles status
  assert_success
  refute_output --partial "CONFLICT PENDING"
}

@test "status: conflict ブランチがあれば警告" {
  create_data_repo
  git -C "$DOTFILES_DIR" branch "conflict/test/20250101-000000"
  run dotfiles status
  assert_success
  assert_output --partial "CONFLICT PENDING"
}
