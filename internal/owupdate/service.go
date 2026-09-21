// Package owupdate is the Overwatch build's own update checker. It points at
// melahoa/accswitcher's GitHub releases rather than the main app's release
// feed (a different repository, numbered independently - see
// buildmode.OverwatchVersion), and unlike the main app's updater it is
// unsigned: releases are still integrity-checked against whatever digest
// GitHub's API supplies, but there is no keypair to manage for every future
// release. The frontend always confirms with the user before installing -
// CheckForUpdate and DownloadAndInstall are deliberately separate calls.
package owupdate

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"

	"TcNo-Acc-Switcher/internal/buildmode"
	"TcNo-Acc-Switcher/internal/security"
)

// Repository is the GitHub repo this build's releases are published to.
const Repository = "melahoa/accswitcher"

const checkTimeout = 15 * time.Second
const installTimeout = 10 * time.Minute

// releaseAssets are this build's own release asset basenames, keyed by
// GOOS/GOARCH - separate from the main app's TcNo-Acc-Switcher.exe.
var releaseAssets = map[string]string{
	"windows/amd64": "BonbonsAccountSwitcher.exe",
}

// AssetMatcher selects this build's release asset for the running platform.
func AssetMatcher(req updater.CheckRequest, assets []github.ReleaseAsset) int {
	want, ok := releaseAssets[req.Platform+"/"+req.Arch]
	if !ok {
		return github.DefaultAssetMatcher(req, assets)
	}
	wantLower := strings.ToLower(want)
	for i, a := range assets {
		if strings.EqualFold(a.Name, want) || strings.ToLower(a.Name) == wantLower {
			return i
		}
	}
	return github.DefaultAssetMatcher(req, assets)
}

// NewProvider builds the GitHub release provider for this build.
func NewProvider() (updater.Provider, error) {
	return github.New(github.Config{
		Repository:   Repository,
		AssetMatcher: AssetMatcher,
	})
}

// UpdateInfoDTO is the JSON shape the frontend checks for an available update.
type UpdateInfoDTO struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Notes     string `json:"notes"`
}

// Service is the Wails-bound entry point the frontend's "Check for updates"
// button (and its own launch-time background check) calls.
type Service struct{}

// ServiceName is the Wails service name.
func (s *Service) ServiceName() string { return "OverwatchUpdateService" }

func runningApp() (*application.App, error) {
	app := application.Get()
	if app == nil {
		return nil, errors.New("application is not ready yet")
	}
	return app, nil
}

// Version returns this running build's own version number, for display
// (e.g. a small version label in the corner of the window) - not to be
// confused with the main app's build/config.yml version.
func (s *Service) Version() string {
	return buildmode.OverwatchVersion
}

// CheckForUpdate asks GitHub for the latest release and reports whether it is
// newer than this running build. It never downloads anything.
func (s *Service) CheckForUpdate() (UpdateInfoDTO, error) {
	if err := security.RequireUnlocked(); err != nil {
		return UpdateInfoDTO{}, err
	}
	app, err := runningApp()
	if err != nil {
		return UpdateInfoDTO{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	release, err := app.Updater.Check(ctx)
	if err != nil {
		return UpdateInfoDTO{}, err
	}
	if release == nil {
		return UpdateInfoDTO{Available: false}, nil
	}
	return UpdateInfoDTO{Available: true, Version: release.Version, Notes: release.Notes}, nil
}

// DownloadAndInstall downloads the release found by the most recent
// CheckForUpdate and restarts the app to apply it. Only call this after the
// user has explicitly confirmed - it is never triggered on its own.
func (s *Service) DownloadAndInstall() error {
	if err := security.RequireUnlocked(); err != nil {
		return err
	}
	app, err := runningApp()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
	defer cancel()
	if err := app.Updater.DownloadAndInstall(ctx); err != nil {
		return err
	}
	return app.Updater.Restart(ctx)
}
