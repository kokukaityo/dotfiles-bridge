setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "link: OS ヘッダーと Done を出力" {
  create_data_repo
  run dotfile link
  assert_success
  assert_output --partial "=== dotfile link"
  assert_output --partial "Done."
}

@test "link: link.toml に従って symlink を作成" {
  can_symlink || skip "symlink creation not available"
  create_data_repo

  local os_key
  os_key="$(get_os_key)"
  local target_file="$HOME/linked-file"

  echo "source content" > "$DOTFILES_DIR/editor/testfile.txt"
  cat > "$DOTFILES_DIR/editor/link.toml" <<TOML
[$os_key]
"testfile.txt" = ["$target_file"]
TOML
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add link config" --quiet

  run dotfile link
  assert_success
  assert_output --partial "linked:"

  run cat "$target_file"
  assert_output "source content"
}

@test "link: 既存ファイルを .bak.* にバックアップ" {
  can_symlink || skip "symlink creation not available"
  create_data_repo

  local os_key
  os_key="$(get_os_key)"
  local target_file="$HOME/existing-file"

  mkdir -p "$(dirname "$target_file")"
  echo "old content" > "$target_file"
  echo "new content" > "$DOTFILES_DIR/editor/testfile.txt"
  cat > "$DOTFILES_DIR/editor/link.toml" <<TOML
[$os_key]
"testfile.txt" = ["$target_file"]
TOML
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add link config" --quiet

  run dotfile link
  assert_success
  assert_output --partial "backed up:"

  local backup_count
  backup_count=$(ls "$HOME"/existing-file.bak.* 2>/dev/null | wc -l)
  [[ "$backup_count" -ge 1 ]] || fail "expected backup file to exist"
}

@test "link: リンク済みならスキップ" {
  can_symlink || skip "symlink creation not available"
  create_data_repo

  local os_key
  os_key="$(get_os_key)"
  local target_file="$HOME/linked-file"

  echo "content" > "$DOTFILES_DIR/editor/testfile.txt"
  cat > "$DOTFILES_DIR/editor/link.toml" <<TOML
[$os_key]
"testfile.txt" = ["$target_file"]
TOML
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add link config" --quiet

  dotfile link

  run dotfile link
  assert_success
  assert_output --partial "already linked"
}

@test "link: ソース不在ならスキップ" {
  create_data_repo

  local os_key
  os_key="$(get_os_key)"
  cat > "$DOTFILES_DIR/editor/link.toml" <<TOML
[$os_key]
"nonexistent.txt" = ["$HOME/target"]
TOML
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "add link config" --quiet

  run dotfile link
  assert_success
  assert_output --partial "skip (source not found)"
}
