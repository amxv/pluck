package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/amxv/pluck/internal/github"
)

type fakeDownloader struct {
	result github.Result
	called int
}

func (f *fakeDownloader) Download(_ context.Context, _ github.Source, _ string, _ bool) (github.Result, error) {
	f.called++
	return f.result, nil
}

func TestRunRootHelp(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	if err := Run([]string{"--help"}, &out, &errBuf); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	for _, expected := range []string{
		"Usage:",
		"github-url",
		"--output",
		"default: /tmp/",
		"/tmp/<source-name>",
		"--force",
		"--version",
		"blob",
	} {
		if !strings.Contains(out.String(), expected) {
			t.Fatalf("expected %q in help output, got: %q", expected, out.String())
		}
	}
}

func TestRunVersion(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	if err := Run([]string{"--version"}, &out, &errBuf); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !strings.HasPrefix(out.String(), "pluck ") {
		t.Fatalf("unexpected version output: %q", out.String())
	}
}

func TestRunDownloadsEverySource(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	fake := &fakeDownloader{result: github.Result{Target: "/tmp/composition-patterns", FileCount: 6}}

	err := run(context.Background(), []string{
		"-o", "/tmp",
		"https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns",
	}, &out, &errBuf, fake)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if fake.called != 1 {
		t.Fatalf("expected one download, got %d", fake.called)
	}
	if !strings.Contains(out.String(), "Saved /tmp/composition-patterns (6 files)") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunDownloadsFileSource(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	fake := &fakeDownloader{result: github.Result{Target: "/tmp/app.go", FileCount: 1}}

	err := run(context.Background(), []string{
		"https://github.com/amxv/go-cli-template/blob/main/internal/app/app.go",
	}, &out, &errBuf, fake)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if fake.called != 1 {
		t.Fatalf("expected one download, got %d", fake.called)
	}
	if !strings.Contains(out.String(), "Saved /tmp/app.go (1 file)") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunRequiresURL(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	err := run(context.Background(), []string{"--output", "/tmp"}, &out, &errBuf, &fakeDownloader{})
	if err == nil || !strings.Contains(err.Error(), "GitHub folder URL") {
		t.Fatalf("expected missing URL error, got %v", err)
	}
}

func TestRunRejectsDuplicateDestinations(t *testing.T) {
	var out bytes.Buffer
	var errBuf bytes.Buffer

	err := run(context.Background(), []string{
		"https://github.com/acme/one/tree/main/shared",
		"https://github.com/acme/two/tree/main/shared",
	}, &out, &errBuf, &fakeDownloader{})
	if err == nil || !strings.Contains(err.Error(), "same destination") {
		t.Fatalf("expected duplicate destination error, got %v", err)
	}
}
