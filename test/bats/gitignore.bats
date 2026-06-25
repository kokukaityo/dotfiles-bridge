setup() {
  load helper/setup
  _common_setup
}

teardown() {
  _common_teardown
}

@test "gitignore: 手書きセクション保持 + 自動セクション再生成" {
  create_data_repo

  cat > "$DOTFILES_DIR/.gitignore" <<'EOF'
my-manual-entry/
another-manual/
# --- auto-generated from sync.toml (do not edit below) ---
old-auto-content
# --- end auto-generated ---
EOF
  git -C "$DOTFILES_DIR" add -A
  git -C "$DOTFILES_DIR" commit -m "gitignore with manual section" --quiet

  run dotfiles gitignore
  assert_success

  assert_file_contains "$DOTFILES_DIR/.gitignore" "my-manual-entry/"
  assert_file_contains "$DOTFILES_DIR/.gitignore" "another-manual/"
  assert_file_contains "$DOTFILES_DIR/.gitignore" "auto-generated from sync.toml"
  assert_file_contains "$DOTFILES_DIR/.gitignore" "backup/"
  assert_file_contains "$DOTFILES_DIR/.gitignore" "*auth*"
  assert_file_contains "$DOTFILES_DIR/.gitignore" ".env*"
  ! grep -qF ".conflict-pending" "$DOTFILES_DIR/.gitignore" || fail "expected .gitignore not to contain .conflict-pending"
  assert_file_contains "$DOTFILES_DIR/.gitignore" ".dotfiles-hook/"
}
