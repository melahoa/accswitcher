package owupdate

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
)

func TestAssetMatcherPicksTheWindowsExeByName(t *testing.T) {
	assets := []github.ReleaseAsset{
		{Name: "SHA256SUMS"},
		{Name: "BonbonsAccountSwitcher.exe"},
		{Name: "BonbonsAccountSwitcher.exe.sig"},
	}
	req := updater.CheckRequest{Platform: "windows", Arch: "amd64"}
	idx := AssetMatcher(req, assets)
	if idx != 1 {
		t.Fatalf("AssetMatcher = %d, want 1 (BonbonsAccountSwitcher.exe)", idx)
	}
}

func TestAssetMatcherIsCaseInsensitive(t *testing.T) {
	assets := []github.ReleaseAsset{{Name: "bonbonsaccountswitcher.EXE"}}
	req := updater.CheckRequest{Platform: "windows", Arch: "amd64"}
	if idx := AssetMatcher(req, assets); idx != 0 {
		t.Fatalf("AssetMatcher = %d, want 0", idx)
	}
}

func TestAssetMatcherFallsBackForUnknownPlatform(t *testing.T) {
	assets := []github.ReleaseAsset{{Name: "app-darwin-arm64.zip"}}
	req := updater.CheckRequest{Platform: "darwin", Arch: "arm64"}
	// No entry in releaseAssets for darwin/arm64 - falls through to Wails'
	// own DefaultAssetMatcher rather than refusing to match anything.
	idx := AssetMatcher(req, assets)
	if idx != github.DefaultAssetMatcher(req, assets) {
		t.Fatalf("AssetMatcher = %d, want DefaultAssetMatcher's result", idx)
	}
}
