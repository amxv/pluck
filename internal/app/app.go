package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/amxv/pluck/internal/buildinfo"
	"github.com/amxv/pluck/internal/github"
)

const commandName = "pluck"

var version = buildinfo.CurrentVersion()

type downloader interface {
	Download(context.Context, github.Source, string, bool) (github.Result, error)
}

func Run(args []string, stdout, stderr io.Writer) error {
	return run(context.Background(), args, stdout, stderr, github.NewClient())
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer, client downloader) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printRootHelp(stdout)
		return nil
	}
	if len(args) == 1 && isVersionArg(args[0]) {
		_, _ = fmt.Fprintf(stdout, "%s %s\n", commandName, version)
		return nil
	}

	flags := flag.NewFlagSet(commandName, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { printRootHelp(stderr) }

	var output string
	flags.StringVar(&output, "output", "/tmp", "destination directory; the source name is appended")
	flags.StringVar(&output, "o", "/tmp", "alias for --output")

	var force bool
	flags.BoolVar(&force, "force", false, "replace an existing extracted folder or file")
	flags.BoolVar(&force, "f", false, "alias for --force")

	if err := flags.Parse(args); err != nil {
		return err
	}
	urls := flags.Args()
	if len(urls) == 0 {
		return errors.New("at least one GitHub folder URL or file URL is required (run `pluck --help`)")
	}
	if strings.TrimSpace(output) == "" {
		return errors.New("output directory cannot be empty")
	}

	sources := make([]github.Source, 0, len(urls))
	seenTargets := make(map[string]string, len(urls))
	for _, rawURL := range urls {
		source, err := github.ParseFolderURL(rawURL)
		if err != nil {
			return fmt.Errorf("%s: %w", rawURL, err)
		}

		target := filepath.Join(output, source.Name())
		if previous, exists := seenTargets[target]; exists {
			return fmt.Errorf("%s and %s resolve to the same destination %q", previous, rawURL, target)
		}
		seenTargets[target] = rawURL
		sources = append(sources, source)
	}

	for _, source := range sources {
		result, err := client.Download(ctx, source, output, force)
		if err != nil {
			return fmt.Errorf("%s: %w", source.URL(), err)
		}
		_, _ = fmt.Fprintf(stdout, "Saved %s (%d file%s)\n", result.Target, result.FileCount, pluralSuffix(result.FileCount))
	}

	return nil
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func isVersionArg(value string) bool {
	return value == "--version"
}

func isHelpArg(value string) bool {
	switch value {
	case "-h", "--help", "help":
		return true
	default:
		return false
	}
}

func printRootHelp(w io.Writer) {
	writeLines(w,
		"pluck - copy a folder or file from a GitHub repository",
		"",
		"Usage:",
		"  pluck [flags] <github-url> [<github-url> ...]",
		"",
		"Flags:",
		"  -o, --output PATH  destination directory (default: /tmp/)",
		"  -f, --force        replace an existing extracted folder or file",
		"      --version     print the installed version",
		"  -h, --help         show this help menu",
		"",
		"By default, each source is saved under /tmp/<source-name>.",
		"The source folder or file name is appended to the destination directory.",
		"GitHub URLs may point to public repositories or repositories available to gh.",
		"",
		"Examples:",
		"  pluck https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns",
		"  pluck https://github.com/amxv/go-cli-template/blob/main/internal/app/app.go",
		"  pluck -o ./skills https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns",
		"  pluck --force -o /tmp URL1 URL2",
	)
}

func writeLines(w io.Writer, lines ...string) {
	for _, line := range lines {
		_, _ = fmt.Fprintln(w, line)
	}
}
