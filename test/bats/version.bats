setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "version: exit 0 で本体 version を出力" {
  run dotfiles version
  assert_success
  assert_output --partial "dotfiles v"
}

@test "version: データリポジトリがあれば data 情報も表示" {
  create_data_repo
  run dotfiles version
  assert_success
  assert_output --partial "data:"
}

@test "version: データリポジトリがなければ本体情報のみ" {
  export DOTFILES_DIR="$TEST_TEMP_DIR/nonexistent"
  run dotfiles version
  assert_success
  refute_output --partial "data:"
}
