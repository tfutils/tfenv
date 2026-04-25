// Package cli provides the command dispatch framework for tfenv.
//
// Subcommands are registered in a map of name to handler function.
// Other packages plug in by adding entries to the registry.
package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Handler is the function signature for a subcommand handler.
// It receives the remaining command-line arguments and returns an exit code.
type Handler func(args []string) int

// command holds metadata for a registered subcommand.
type command struct {
	handler     Handler
	description string
}

// registry maps subcommand names to their handlers and descriptions.
var registry = map[string]command{}

// Register adds a subcommand to the dispatch registry.
func Register(name string, description string, handler Handler) {
	registry[name] = command{
		handler:     handler,
		description: description,
	}
}

// BuildInfo holds version metadata injected at build time.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// Run dispatches to the appropriate subcommand based on args.
// It returns an exit code suitable for os.Exit.
func Run(info BuildInfo, args []string) int {
	if len(args) == 0 {
		printUsage(info.Version)
		return 0
	}

	subcmd := args[0]

	// Handle --version and version as special cases.
	if subcmd == "--version" || subcmd == "version" {
		fmt.Fprintf(os.Stdout, "tfenv %s (commit: %s, built: %s)\n", info.Version, info.Commit, info.Date)
		return 0
	}

	// Handle help as a special case.
	if subcmd == "help" || subcmd == "--help" || subcmd == "-h" {
		printUsage(info.Version)
		return 0
	}

	// Look up the subcommand in the registry.
	cmd, ok := registry[subcmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "tfenv: unknown command %q\n", subcmd)
		fmt.Fprintf(os.Stderr, "Run 'tfenv help' for usage.\n")
		return 1
	}

	return cmd.handler(args[1:])
}

// printUsage prints the help text listing all registered subcommands.
func printUsage(version string) {
	fmt.Fprintf(os.Stdout, "tfenv %s\n\n", version)
	fmt.Fprintf(os.Stdout, "Usage: tfenv <command> [args]\n\n")

	// Always include the built-in commands.
	builtins := []struct {
		name string
		desc string
	}{
		{"help", "Show this help output"},
		{"version", "Print tfenv version"},
	}

	// Collect registered commands and sort them.
	var names []string
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Fprintf(os.Stdout, "Commands:\n")

	// Print built-in commands first.
	for _, b := range builtins {
		fmt.Fprintf(os.Stdout, "  %-16s %s\n", b.name, b.desc)
	}

	// Print registered commands.
	for _, name := range names {
		cmd := registry[name]
		fmt.Fprintf(os.Stdout, "  %-16s %s\n", name, cmd.description)
	}

	// Build the full list for "Available commands" summary.
	var all []string
	for _, b := range builtins {
		all = append(all, b.name)
	}
	all = append(all, names...)
	sort.Strings(all)

	fmt.Fprintf(os.Stdout, "\nAvailable commands: %s\n", strings.Join(all, ", "))
}
