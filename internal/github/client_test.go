package github

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFolderURL(t *testing.T) {
	source, err := ParseFolderURL("https://github.com/vercel-labs/agent-skills/tree/main/skills/composition-patterns")
	if err != nil {
		t.Fatalf("ParseFolderURL returned error: %v", err)
	}
	if source.Owner != "vercel-labs" || source.Repo != "agent-skills" {
		t.Fatalf("unexpected repository: %#v", source)
	}
	if source.Name() != "composition-patterns" {
		t.Fatalf("unexpected folder name: %q", source.Name())
	}
}

func TestParseFolderURLRejectsNonFolderURLs(t *testing.T) {
	for _, rawURL := range []string{
		"https://github.com/vercel-labs/agent-skills",
		"https://github.com/vercel-labs/agent-skills/blob/main/README.md",
		"http://github.com/vercel-labs/agent-skills/tree/main/skills",
		"https://example.com/vercel-labs/agent-skills/tree/main/skills",
	} {
		if _, err := ParseFolderURL(rawURL); err == nil {
			t.Errorf("expected %q to be rejected", rawURL)
		}
	}
}

func TestDownloadFolder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/repos/acme/demo/commits/main":
			_, _ = writer.Write([]byte(`{"sha":"commit-sha"}`))
		case "/repos/acme/demo/git/trees/commit-sha":
			_, _ = writer.Write([]byte(`{
                "truncated": false,
                "tree": [
                    {"path":"docs","mode":"040000","type":"tree","sha":"docs-sha"},
                    {"path":"docs/README.md","mode":"100644","type":"blob","sha":"readme-sha"},
                    {"path":"docs/tools","mode":"040000","type":"tree","sha":"tools-sha"},
                    {"path":"docs/tools/run.sh","mode":"100755","type":"blob","sha":"run-sha"}
                ]
            }`))
		case "/repos/acme/demo/git/blobs/readme-sha":
			writeBlobResponse(writer, []byte("hello\n"))
		case "/repos/acme/demo/git/blobs/run-sha":
			writeBlobResponse(writer, []byte("#!/bin/sh\necho hi\n"))
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: server.Client(),
		apiBaseURL: server.URL,
		userAgent:  "pluck-test",
	}
	source, err := ParseFolderURL("https://github.com/acme/demo/tree/main/docs")
	if err != nil {
		t.Fatalf("ParseFolderURL returned error: %v", err)
	}

	result, err := client.Download(t.Context(), source, t.TempDir(), false)
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if result.FileCount != 2 {
		t.Fatalf("expected 2 files, got %d", result.FileCount)
	}

	readme, err := os.ReadFile(filepath.Join(result.Target, "README.md"))
	if err != nil {
		t.Fatalf("read extracted README: %v", err)
	}
	if string(readme) != "hello\n" {
		t.Fatalf("unexpected README: %q", readme)
	}
	info, err := os.Stat(filepath.Join(result.Target, "tools", "run.sh"))
	if err != nil {
		t.Fatalf("stat extracted script: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("expected executable mode, got %o", info.Mode().Perm())
	}
}

func TestDownloadRejectsExistingDestination(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	client := &Client{httpClient: server.Client(), apiBaseURL: server.URL}
	source, err := ParseFolderURL("https://github.com/acme/demo/tree/main/docs")
	if err != nil {
		t.Fatalf("ParseFolderURL returned error: %v", err)
	}
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err = client.Download(t.Context(), source, root, false)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected existing destination error, got %v", err)
	}
}

func writeBlobResponse(writer http.ResponseWriter, contents []byte) {
	payload := map[string]string{
		"encoding": "base64",
		"content":  base64.StdEncoding.EncodeToString(contents),
	}
	_ = json.NewEncoder(writer).Encode(payload)
}
