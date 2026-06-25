setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "version: exit 0 で engine version を出力" {
  run dotfile version
  assert_success
  assert_output --partial "dotfile engine v"
}

@test "version: データリポジトリがあれば data 情報も表示" {
  create_data_repo
  run dotfile version
  assert_success
  assert_output --partial "data:"
  assert_output --partial "data version:"
}

@test "version: データリポジトリがなければ engine 情報のみ" {
  export DOTFILES_DIR="$TEST_TEMP_DIR/nonexistent"
  run dotfile version
  assert_success
  refute_output --partial "data:"
}
