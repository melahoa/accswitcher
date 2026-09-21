// Package owupdate is the Overwatch build's own update checker. It points at
// melahoa/accswitcher's GitHub releases rather than the main app's release
// feed (a different repository, numbered independently - see
// buildmode.OverwatchVersion).
//
// Deliberately a plain read-only HTTP GET against GitHub's public REST API,
// not the Wails updater package's self-download-and-replace machinery: a
// process that downloads an executable, overwrites itself, and relaunches is
// close to a textbook heuristic signature for dropper/trojan malware, and an
// earlier version of this file did exactly that using the Wails updater -
// which got this build flagged by Windows Defender. Checking is the only
// thing this package does; installing an update is always a normal manual
// download the user drives themselves, in their own browser.
package owupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"TcNo-Acc-Switcher/internal/buildmode"
	"TcNo-Acc-Switcher/internal/security"
)

// Repository is the GitHub repo this build's releases are published to.
const Repository = "melahoa/accswitcher"

const checkTimeout = 15 * time.Second

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Body    string `json:"body"`
}

// UpdateInfoDTO is the JSON shape the frontend checks for an available update.
type UpdateInfoDTO struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Notes     string `json:"notes"`
	// URL is the release page to open in the user's browser - this build
	// never downloads or applies anything on its own.
	URL string `json:"url"`
}

// Service is the Wails-bound entry point the frontend's "Check for updates"
// button (and its own launch-time background check) calls.
type Service struct{}

// ServiceName is the Wails service name.
func (s *Service) ServiceName() string { return "OverwatchUpdateService" }

// Version returns this running build's own version number, for display
// (e.g. a small version label in the corner of the window) - not to be
// confused with the main app's build/config.yml version.
func (s *Service) Version() string {
	return buildmode.OverwatchVersion
}

// CheckForUpdate asks GitHub for the latest release and reports whether it is
// newer than this running build. It never downloads anything - installing is
// always a manual step the user takes in their own browser.
func (s *Service) CheckForUpdate() (UpdateInfoDTO, error) {
	if err := security.RequireUnlocked(); err != nil {
		return UpdateInfoDTO{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://api.github.com/repos/"+Repository+"/releases/latest", nil)
	if err != nil {
		return UpdateInfoDTO{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "BonbonsAccountSwitcher/"+buildmode.OverwatchVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UpdateInfoDTO{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return UpdateInfoDTO{}, fmt.Errorf("owupdate: GitHub API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return UpdateInfoDTO{}, err
	}
	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return UpdateInfoDTO{}, err
	}

	latest := strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	if !isNewer(latest, buildmode.OverwatchVersion) {
		return UpdateInfoDTO{Available: false}, nil
	}
	return UpdateInfoDTO{
		Available: true,
		Version:   latest,
		Notes:     release.Body,
		URL:       release.HTMLURL,
	}, nil
}

// isNewer does a plain MAJOR.MINOR.PATCH numeric comparison. No prerelease or
// build-metadata handling - this build's own tags are always plain vX.Y.Z.
func isNewer(latest, current string) bool {
	l := parseVersion(latest)
	c := parseVersion(current)
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	var out [3]int
	parts := strings.SplitN(v, ".", 3)
	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(strings.TrimSpace(parts[i]))
		out[i] = n
	}
	return out
}
