package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const defaultAPIBaseURL = "https://api.github.com"

// Source identifies a folder URL on github.com. The path after /tree/ is
// kept as segments so branch names containing slashes can be resolved.
type Source struct {
	Owner string
	Repo  string
	parts []string
	raw   string
}

func (s Source) Name() string {
	if len(s.parts) == 0 {
		return ""
	}
	return s.parts[len(s.parts)-1]
}

func (s Source) URL() string {
	return s.raw
}

// Result describes one successfully extracted source folder.
type Result struct {
	Target    string
	FileCount int
}

// Client downloads GitHub repository trees and blobs through the GitHub API.
type Client struct {
	httpClient httpClient
	apiBaseURL string
	token      string
	userAgent  string
}

type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

// NewClient creates a GitHub API client. It uses GH_TOKEN or GITHUB_TOKEN
// when set, then falls back to the active gh CLI account for private repos.
func NewClient() *Client {
	token := strings.TrimSpace(os.Getenv("GH_TOKEN"))
	if token == "" {
		token = strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	}
	if token == "" {
		token = ghAuthToken()
	}

	return &Client{
		httpClient: http.DefaultClient,
		apiBaseURL: defaultAPIBaseURL,
		token:      token,
		userAgent:  "pluck",
	}
}

// ParseFolderURL validates a GitHub tree URL and keeps its repository path.
func ParseFolderURL(raw string) (Source, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Source{}, errors.New("URL cannot be empty")
	}

	u, err := url.Parse(trimmed)
	if err != nil {
		return Source{}, fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" || strings.ToLower(u.Hostname()) != "github.com" {
		return Source{}, errors.New("URL must use https://github.com")
	}
	if u.User != nil {
		return Source{}, errors.New("URL must not include user information")
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return Source{}, errors.New("URL must not include a query string or fragment")
	}

	parts, err := pathSegments(u.EscapedPath())
	if err != nil {
		return Source{}, err
	}
	if len(parts) < 5 || parts[2] != "tree" {
		return Source{}, errors.New("URL must point to a folder using /owner/repo/tree/ref/path")
	}
	if parts[0] == "" || parts[1] == "" {
		return Source{}, errors.New("URL is missing a repository owner or name")
	}
	if len(parts[3:]) < 2 {
		return Source{}, errors.New("URL must include a folder path after the branch or tag")
	}

	return Source{
		Owner: parts[0],
		Repo:  strings.TrimSuffix(parts[1], ".git"),
		parts: parts[3:],
		raw:   trimmed,
	}, nil
}

func pathSegments(rawPath string) ([]string, error) {
	trimmed := strings.Trim(rawPath, "/")
	if trimmed == "" {
		return nil, errors.New("URL path cannot be empty")
	}

	rawParts := strings.Split(trimmed, "/")
	parts := make([]string, 0, len(rawParts))
	for _, rawPart := range rawParts {
		part, err := url.PathUnescape(rawPart)
		if err != nil {
			return nil, fmt.Errorf("invalid escaped path segment: %w", err)
		}
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, "/\\\x00") {
			return nil, errors.New("URL contains an invalid path segment")
		}
		parts = append(parts, part)
	}
	return parts, nil
}

// Download extracts the source folder into outputRoot/source.Name().
func (c *Client) Download(ctx context.Context, source Source, outputRoot string, force bool) (Result, error) {
	if c == nil || c.httpClient == nil {
		return Result{}, errors.New("GitHub client is not configured")
	}
	if len(source.parts) < 2 {
		return Result{}, errors.New("source URL does not include a folder path")
	}

	root, err := filepath.Abs(outputRoot)
	if err != nil {
		return Result{}, fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory: %w", err)
	}

	target := filepath.Join(root, source.Name())
	if !isWithin(root, target) {
		return Result{}, errors.New("source folder name would escape the output directory")
	}
	if _, err := os.Lstat(target); err == nil && !force {
		return Result{}, fmt.Errorf("destination %q already exists (use --force to replace it)", target)
	} else if err != nil && !os.IsNotExist(err) {
		return Result{}, fmt.Errorf("inspect destination: %w", err)
	}

	entries, err := c.folderEntries(ctx, source)
	if err != nil {
		return Result{}, err
	}

	staging, err := os.MkdirTemp(root, "."+source.Name()+".tmp-")
	if err != nil {
		return Result{}, fmt.Errorf("create staging directory: %w", err)
	}
	defer os.RemoveAll(staging)

	fileCount, err := c.materialize(ctx, staging, source, entries)
	if err != nil {
		return Result{}, err
	}

	if force {
		if err := os.RemoveAll(target); err != nil {
			return Result{}, fmt.Errorf("replace destination: %w", err)
		}
	}
	if err := os.Rename(staging, target); err != nil {
		return Result{}, fmt.Errorf("finalize destination: %w", err)
	}

	return Result{Target: target, FileCount: fileCount}, nil
}

type treeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
}

type treeResponse struct {
	Truncated bool        `json:"truncated"`
	Tree      []treeEntry `json:"tree"`
}

func (c *Client) folderEntries(ctx context.Context, source Source) ([]treeEntry, error) {
	_, folder, commitSHA, err := c.resolveSource(ctx, source)
	if err != nil {
		return nil, err
	}

	var response treeResponse
	endpoint := fmt.Sprintf(
		"/repos/%s/%s/git/trees/%s?recursive=1",
		url.PathEscape(source.Owner),
		url.PathEscape(source.Repo),
		url.PathEscape(commitSHA),
	)
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return nil, fmt.Errorf("read repository tree: %w", err)
	}
	if response.Truncated {
		return nil, errors.New("GitHub returned a truncated repository tree; the folder is too large to copy safely")
	}

	prefix := folder + "/"
	foundFolder := false
	entries := make([]treeEntry, 0)
	for _, entry := range response.Tree {
		if entry.Path == folder {
			if entry.Type != "tree" {
				return nil, fmt.Errorf("path %q is not a folder", folder)
			}
			foundFolder = true
			continue
		}
		if !strings.HasPrefix(entry.Path, prefix) {
			continue
		}

		relative := strings.TrimPrefix(entry.Path, prefix)
		if relative == "" {
			continue
		}
		entry.Path = relative
		entries = append(entries, entry)
		foundFolder = true
	}
	if !foundFolder {
		return nil, fmt.Errorf("folder %q was not found in the repository", folder)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func (c *Client) resolveSource(ctx context.Context, source Source) (string, string, string, error) {
	var notFound error
	for refParts := 1; refParts < len(source.parts); refParts++ {
		ref := strings.Join(source.parts[:refParts], "/")
		folder := strings.Join(source.parts[refParts:], "/")
		commitSHA, err := c.resolveCommit(ctx, source.Owner, source.Repo, ref)
		if err == nil {
			return ref, folder, commitSHA, nil
		}

		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Status != http.StatusNotFound {
			return "", "", "", err
		}
		notFound = err
	}

	if notFound != nil {
		return "", "", "", fmt.Errorf("could not resolve branch or tag in URL: %w", notFound)
	}
	return "", "", "", errors.New("URL does not include a branch or tag")
}

func (c *Client) resolveCommit(ctx context.Context, owner, repo, ref string) (string, error) {
	var response struct {
		SHA string `json:"sha"`
	}
	endpoint := fmt.Sprintf(
		"/repos/%s/%s/commits/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.PathEscape(ref),
	)
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.SHA) == "" {
		return "", errors.New("GitHub returned an empty commit SHA")
	}
	return response.SHA, nil
}

func (c *Client) materialize(ctx context.Context, staging string, source Source, entries []treeEntry) (int, error) {
	fileCount := 0
	for _, entry := range entries {
		destination, err := safeJoin(staging, entry.Path)
		if err != nil {
			return 0, err
		}

		switch entry.Type {
		case "tree":
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return 0, fmt.Errorf("create directory %q: %w", entry.Path, err)
			}
		case "blob":
			contents, err := c.getBlob(ctx, source.Owner, source.Repo, entry.SHA)
			if err != nil {
				return 0, fmt.Errorf("download %q: %w", entry.Path, err)
			}
			if err := writeBlob(destination, contents, entry.Mode); err != nil {
				return 0, fmt.Errorf("write %q: %w", entry.Path, err)
			}
			fileCount++
		case "commit":
			return 0, fmt.Errorf("submodule %q is not supported", entry.Path)
		default:
			return 0, fmt.Errorf("GitHub returned unsupported tree entry type %q for %q", entry.Type, entry.Path)
		}
	}
	return fileCount, nil
}

func (c *Client) getBlob(ctx context.Context, owner, repo, sha string) ([]byte, error) {
	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	endpoint := fmt.Sprintf(
		"/repos/%s/%s/git/blobs/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		url.PathEscape(sha),
	)
	if err := c.getJSON(ctx, endpoint, &response); err != nil {
		return nil, err
	}
	if response.Encoding != "base64" {
		return nil, fmt.Errorf("GitHub returned unsupported blob encoding %q", response.Encoding)
	}
	encoded := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			return -1
		}
		return r
	}, response.Content)
	contents, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode GitHub blob: %w", err)
	}
	return contents, nil
}

func writeBlob(destination string, contents []byte, mode string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	if mode == "120000" {
		linkTarget := strings.TrimSuffix(string(contents), "\n")
		if err := validateSymlinkTarget(linkTarget); err != nil {
			return err
		}
		return os.Symlink(linkTarget, destination)
	}
	return os.WriteFile(destination, contents, filePermissions(mode))
}

func filePermissions(mode string) os.FileMode {
	parsed, err := strconv.ParseUint(mode, 8, 32)
	if err != nil {
		return 0o644
	}
	permissions := os.FileMode(parsed & 0o777)
	if permissions == 0 {
		return 0o644
	}
	return permissions
}

func validateSymlinkTarget(target string) error {
	if target == "" || strings.ContainsAny(target, "\x00\\") || filepath.IsAbs(target) {
		return errors.New("repository symlink has an unsafe target")
	}
	cleaned := path.Clean(target)
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return errors.New("repository symlink points outside the extracted folder")
	}
	return nil
}

func safeJoin(root, relative string) (string, error) {
	if relative == "" || strings.ContainsAny(relative, "\x00\\") {
		return "", fmt.Errorf("unsafe repository path %q", relative)
	}
	cleaned := path.Clean(relative)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("unsafe repository path %q", relative)
	}

	destination := filepath.Join(root, filepath.FromSlash(cleaned))
	if !isWithin(root, destination) {
		return "", fmt.Errorf("repository path %q escapes the output directory", relative)
	}
	return destination, nil
}

func isWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func (c *Client) getJSON(ctx context.Context, endpoint string, destination any) error {
	requestURL := strings.TrimRight(c.apiBaseURL, "/") + endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return fmt.Errorf("create GitHub request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request GitHub: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 32*1024))
		message := strings.TrimSpace(string(body))
		var payload struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &payload) == nil && payload.Message != "" {
			message = payload.Message
		}
		if message == "" {
			message = http.StatusText(response.StatusCode)
		}
		return &APIError{Status: response.StatusCode, Message: message}
	}

	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode GitHub response: %w", err)
	}
	return nil
}

// APIError is returned for non-successful GitHub API responses.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("GitHub API returned HTTP %d: %s", e.Status, e.Message)
}

func ghAuthToken() string {
	command := exec.Command("gh", "auth", "token")
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
