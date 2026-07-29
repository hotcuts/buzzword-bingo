package updatecheck

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bingo/internal/config"
)

const (
	Repo       = "hotcuts/buzzword-bingo"
	cacheFile  = "update-check.json"
	defaultTTL = 24 * time.Hour
	timeout    = 1500 * time.Millisecond
)

// Checker performs a best-effort latest-release lookup and prints a suggestion
// when the remote tag is newer than the running binary.
type Checker struct {
	Client     *http.Client
	APIBaseURL string
	CachePath  string
	Out        io.Writer
	Now        func() time.Time
	CacheTTL   time.Duration
}

type cacheFileData struct {
	CheckedAt string `json:"checked_at"`
	LatestTag string `json:"latest_tag"`
}

// MaybeNotify starts a silent, best-effort update check for currentVersion
// using the default config dir and GitHub Releases API.
func MaybeNotify(currentVersion string) {
	cfg, err := config.Ensure()
	if err != nil {
		return
	}
	c := &Checker{
		CachePath: filepath.Join(cfg.Dir, cacheFile),
		Out:       os.Stdout,
	}
	c.MaybeNotify(currentVersion)
}

// MaybeNotify checks for a newer release and prints a one-line suggestion when
// found. Network and I/O errors are ignored.
func (c *Checker) MaybeNotify(currentVersion string) {
	if c == nil {
		return
	}
	out := c.Out
	if out == nil {
		out = os.Stdout
	}
	nowFn := c.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	now := nowFn()
	ttl := c.CacheTTL
	if ttl == 0 {
		ttl = defaultTTL
	}

	latest, fromCache := c.latestFromCache(now, ttl)
	if !fromCache {
		tag, err := c.fetchLatest()
		if err != nil {
			return
		}
		latest = tag
		_ = c.writeCache(now, latest)
	}

	if !IsNewer(latest, currentVersion) {
		return
	}
	_, _ = fmt.Fprintf(out, "Update available: %s (you have %s). Run: bingo update\n",
		displayTag(latest), displayTag(currentVersion))
}

func (c *Checker) latestFromCache(now time.Time, ttl time.Duration) (string, bool) {
	if c.CachePath == "" {
		return "", false
	}
	raw, err := os.ReadFile(c.CachePath)
	if err != nil {
		return "", false
	}
	var data cacheFileData
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", false
	}
	if data.LatestTag == "" {
		return "", false
	}
	checkedAt, err := time.Parse(time.RFC3339, data.CheckedAt)
	if err != nil {
		return "", false
	}
	if now.Sub(checkedAt) > ttl {
		return "", false
	}
	return data.LatestTag, true
}

func (c *Checker) writeCache(now time.Time, latestTag string) error {
	if c.CachePath == "" {
		return nil
	}
	data, err := json.Marshal(cacheFileData{
		CheckedAt: now.UTC().Format(time.RFC3339),
		LatestTag: latestTag,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(c.CachePath, data, 0600)
}

func (c *Checker) fetchLatest() (string, error) {
	base := c.APIBaseURL
	if base == "" {
		base = "https://api.github.com"
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	url := strings.TrimRight(base, "/") + "/repos/" + Repo + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "bingo-updatecheck")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}
	var body struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.TagName == "" {
		return "", fmt.Errorf("empty tag")
	}
	return body.TagName, nil
}

// IsNewer reports whether remote is a strictly newer X.Y.Z than local.
// Tags may include a leading "v". Unparseable versions return false.
func IsNewer(remote, local string) bool {
	r, okR := parseSemver(remote)
	l, okL := parseSemver(local)
	if !okR || !okL {
		return false
	}
	for i := 0; i < 3; i++ {
		if r[i] > l[i] {
			return true
		}
		if r[i] < l[i] {
			return false
		}
	}
	return false
}

func parseSemver(v string) ([3]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}

func displayTag(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return v
	}
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}
