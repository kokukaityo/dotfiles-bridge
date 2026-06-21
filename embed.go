package dotfile

import "embed"

//go:embed all:template
var TemplateFS embed.FS

//go:embed VERSION
var Version string

//go:embed lib/hooks/pre-push lib/hooks/post-merge
var HookFS embed.FS
