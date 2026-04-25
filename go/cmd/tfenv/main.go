// Package main provides the multi-call entry point for the tfenv Go edition.
//
// When invoked as "tfenv", it dispatches to the CLI subcommand handler.
// When invoked as "terraform" (e.g. via symlink), it delegates to the shim.
//
// Multi-call detection uses the raw basename of os.Args[0] (before symlink
// resolution). A symlink named "terraform" → "tfenv" will see "terraform"
// as the basename and route to the shim.
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/tfutils/tfenv/go/internal/cli"
	"github.com/tfutils/tfenv/go/internal/shim"

	// Register subcommands via init().
	_ "github.com/tfutils/tfenv/go/internal/install"
	_ "github.com/tfutils/tfenv/go/internal/list"
	_ "github.com/tfutils/tfenv/go/internal/pin"
	_ "github.com/tfutils/tfenv/go/internal/uninstall"
	_ "github.com/tfutils/tfenv/go/internal/use"
)

// version is set at build time via -ldflags "-X main.version=...".
// It defaults to "dev" for local builds.
var version = "dev"

func main() {
	basename := filepath.Base(os.Args[0])

	// Strip .exe suffix on Windows for multi-call matching.
	if runtime.GOOS == "windows" {
		basename = strings.TrimSuffix(basename, ".exe")
	}

	switch basename {
	case "terraform":
		os.Exit(shim.Run(os.Args[1:]))
	default:
		os.Exit(cli.Run(version, os.Args[1:]))
	}
}
