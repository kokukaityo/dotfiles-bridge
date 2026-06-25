setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "pull: リモート先行なら fast-forward" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  local second="$TEST_TEMP_DIR/second"
  git clone "$bare" "$second" --quiet
  git -C "$second" config user.name "bats-test"
  git -C "$second" config user.email "bats@test.invalid"
  mkdir -p "$second/editor"
  echo "remote change" > "$second/editor/remote-file.txt"
  git -C "$second" add -A
  git -C "$second" commit -m "remote update" --quiet
  git -C "$second" push origin main --quiet 2>/dev/null

  run dotfiles pull
  assert_success
  assert_output --partial "Fast-forwarded"
  [[ -f "$DOTFILES_DIR/editor/remote-file.txt" ]] || fail "expected remote file to exist locally"
}

@test "pull: 最新なら何もしない" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  run dotfiles pull
  assert_success
  assert_output --partial "Already up to date."
}

@test "pull: 分岐したら退避ブランチ作成" {
  create_data_repo
  local bare="$TEST_TEMP_DIR/remote.git"
  create_bare_remote "$bare"
  add_remote_to_repo "$DOTFILES_DIR" "$bare" "main"
  enable_remote_mode

  echo "local change" > "$DOTFILES_DIR/editor/local.txt"
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "local commit" --quiet

  local second="$TEST_TEMP_DIR/second"
  git clone "$bare" "$second" --quiet
  git -C "$second" config user.name "bats-test"
  git -C "$second" config user.email "bats@test.invalid"
  mkdir -p "$second/editor"
  echo "remote diverge" > "$second/editor/remote.txt"
  git -C "$second" add -A
  git -C "$second" commit -m "remote diverge" --quiet
  git -C "$second" push origin main --quiet 2>/dev/null

  run dotfiles pull
  assert_success
  assert_output --partial "分岐を検出"

  run git -C "$DOTFILES_DIR" branch --list "conflict/*"
  [[ -n "$output" ]] || fail "expected conflict/* branch to exist"
}
