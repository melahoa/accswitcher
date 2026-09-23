// Package owrank stores each account's manually-entered Overwatch competitive
// standing - one rank per role, since a single account routinely carries a
// different rank in each of Tank, Damage, Support and Open Queue. There is no
// game API for this: ranks only ever come from the player typing them in, so
// the store is a flat JSON file rather than anything fetched or swept.
package owrank

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"TcNo-Acc-Switcher/internal/fsutil"
	"TcNo-Acc-Switcher/internal/paths"
)

const (
	Version      = 1
	maxFileBytes = 256 << 10
	maxEntries   = 4096
	fileName     = "OverwatchRanks.json"
)

// Role is one of the four Overwatch role queues.
type Role string

const (
	RoleTank    Role = "tank"
	RoleDamage  Role = "damage"
	RoleSupport Role = "support"
	RoleOpen    Role = "open"
)

// Roles lists the four roles in the fixed order the UI displays them in.
var Roles = []Role{RoleTank, RoleDamage, RoleSupport, RoleOpen}

func ValidRole(r Role) bool {
	switch r {
	case RoleTank, RoleDamage, RoleSupport, RoleOpen:
		return true
	default:
		return false
	}
}

// Tiers is the competitive ladder, lowest first. Emerald sits between
// Platinum and Diamond per the current (2026 Season 4) ladder; Top 500 is a
// leaderboard placement rather than a rank, so it has no tier here.
var Tiers = []string{
	"bronze", "silver", "gold", "platinum", "emerald", "diamond", "master", "grandmaster", "champion",
}

// TierIndex returns a tier's position on the ladder, or -1 if it is not one of
// [Tiers].
func TierIndex(tier string) int {
	for i, t := range Tiers {
		if t == tier {
			return i
		}
	}
	return -1
}

func ValidTier(tier string) bool { return TierIndex(tier) >= 0 }

const (
	MinDivision = 1
	MaxDivision = 5
)

// RoleRank is one role's standing. The zero value (empty Tier) means the role
// has not been ranked, which is distinct from - and sorts below - Bronze 5.
type RoleRank struct {
	Tier     string `json:"tier"`
	Division int    `json:"division"`
}

func (r RoleRank) Ranked() bool { return r.Tier != "" }

func (r RoleRank) valid() bool {
	if r.Tier == "" {
		return r.Division == 0
	}
	return ValidTier(r.Tier) && r.Division >= MinDivision && r.Division <= MaxDivision
}

// Score orders ranks for sorting: higher is better, unranked sorts lowest of
// all. Division counts down as a player climbs (division 1 beats division 5
// within a tier), so it is inverted before stacking on top of the tier's
// block - otherwise Bronze 1 would outscore Silver 5.
func (r RoleRank) Score() int {
	idx := TierIndex(r.Tier)
	if idx < 0 {
		return -1
	}
	div := r.Division
	if div < MinDivision || div > MaxDivision {
		div = MaxDivision
	}
	return idx*MaxDivision + (MaxDivision - div)
}

// Entry is one account's standings across all four roles, plus whether the
// Overwatch build's own list should show it and pin it to the top. Neither
// flag ever touches the account's real login files or the platform's own
// account list - they only affect this build's own view of it.
type Entry struct {
	PlatformKey string            `json:"platformKey"`
	UniqueID    string            `json:"uniqueId"`
	Roles       map[Role]RoleRank `json:"roles"`
	Hidden      bool              `json:"hidden"`
	Favorite    bool              `json:"favorite"`
	UpdatedAt   int64             `json:"updatedAt"`
}

func key(platformKey, uniqueID string) string {
	return strings.TrimSpace(platformKey) + "|" + strings.TrimSpace(uniqueID)
}

var (
	ErrInvalidStore = errors.New("invalid Overwatch rank store")

	// writeMu serializes load-mutate-save: the editor can save while a second
	// window (or a second account's save) is in flight.
	writeMu sync.Mutex
)

type storeFile struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

// Path returns the on-disk location of the rank store. It lives under the
// shared Settings directory, not a per-platform login cache, since a single
// file spans both Steam and Battle.net accounts.
func Path() (string, error) {
	dir, err := paths.SettingsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

// Load returns every stored entry, keyed by "<platformKey>|<uniqueId>". A
// missing file is the normal state before any rank has ever been set.
func Load() (map[string]Entry, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]Entry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	raw, err := io.ReadAll(io.LimitReader(f, maxFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxFileBytes {
		return nil, ErrInvalidStore
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var stored storeFile
	if err := dec.Decode(&stored); err != nil {
		return nil, errors.Join(ErrInvalidStore, err)
	}
	if dec.Decode(&struct{}{}) != io.EOF || stored.Version != Version || len(stored.Entries) > maxEntries {
		return nil, ErrInvalidStore
	}
	out := make(map[string]Entry, len(stored.Entries))
	for _, entry := range stored.Entries {
		if entry.PlatformKey == "" || entry.UniqueID == "" || entry.UpdatedAt < 0 {
			return nil, ErrInvalidStore
		}
		for role, rr := range entry.Roles {
			if !ValidRole(role) || !rr.valid() {
				return nil, ErrInvalidStore
			}
		}
		k := key(entry.PlatformKey, entry.UniqueID)
		if _, exists := out[k]; exists {
			return nil, ErrInvalidStore
		}
		out[k] = entry
	}
	return out, nil
}

// Get reads one account's ranks. A missing entry is not an error - it just
// means every role is unranked.
func Get(platformKey, uniqueID string) (Entry, error) {
	entries, err := Load()
	if err != nil {
		return Entry{}, err
	}
	if entry, ok := entries[key(platformKey, uniqueID)]; ok {
		return entry, nil
	}
	return Entry{PlatformKey: platformKey, UniqueID: uniqueID, Roles: map[Role]RoleRank{}}, nil
}

// Put replaces all four roles for an account at once: the editor always sends
// every slot together, so there is no separate per-role patch to reconcile.
// An unranked role (empty Tier) is dropped rather than stored, keeping the
// file free of Bronze-1-shaped placeholders for roles nobody has set.
func Put(platformKey, uniqueID string, roles map[Role]RoleRank, now time.Time) error {
	platformKey = strings.TrimSpace(platformKey)
	uniqueID = strings.TrimSpace(uniqueID)
	if platformKey == "" || uniqueID == "" {
		return ErrInvalidStore
	}
	clean := make(map[Role]RoleRank, len(roles))
	for role, rr := range roles {
		if !ValidRole(role) {
			return ErrInvalidStore
		}
		if !rr.valid() {
			return ErrInvalidStore
		}
		if rr.Ranked() {
			clean[role] = rr
		}
	}

	writeMu.Lock()
	defer writeMu.Unlock()
	entries, err := Load()
	if err != nil {
		return err
	}
	// Preserve Hidden/Favorite: editing ranks is independent of hiding or
	// favoriting an account, and a rank save should never silently change
	// either.
	existing := entries[key(platformKey, uniqueID)]
	entries[key(platformKey, uniqueID)] = Entry{
		PlatformKey: platformKey,
		UniqueID:    uniqueID,
		Roles:       clean,
		Hidden:      existing.Hidden,
		Favorite:    existing.Favorite,
		UpdatedAt:   now.Unix(),
	}
	return save(entries)
}

// SetHidden shows or hides an account in the Overwatch build's own list,
// leaving its ranks, and everything the platform itself knows about the
// account, untouched.
func SetHidden(platformKey, uniqueID string, hidden bool, now time.Time) error {
	platformKey = strings.TrimSpace(platformKey)
	uniqueID = strings.TrimSpace(uniqueID)
	if platformKey == "" || uniqueID == "" {
		return ErrInvalidStore
	}

	writeMu.Lock()
	defer writeMu.Unlock()
	entries, err := Load()
	if err != nil {
		return err
	}
	k := key(platformKey, uniqueID)
	entry := entries[k]
	entry.PlatformKey = platformKey
	entry.UniqueID = uniqueID
	entry.Hidden = hidden
	entry.UpdatedAt = now.Unix()
	entries[k] = entry
	return save(entries)
}

// SetFavorite pins or unpins an account to the top of the Overwatch build's
// own list, leaving its ranks, hidden state, and everything the platform
// itself knows about the account, untouched.
func SetFavorite(platformKey, uniqueID string, favorite bool, now time.Time) error {
	platformKey = strings.TrimSpace(platformKey)
	uniqueID = strings.TrimSpace(uniqueID)
	if platformKey == "" || uniqueID == "" {
		return ErrInvalidStore
	}

	writeMu.Lock()
	defer writeMu.Unlock()
	entries, err := Load()
	if err != nil {
		return err
	}
	k := key(platformKey, uniqueID)
	entry := entries[k]
	entry.PlatformKey = platformKey
	entry.UniqueID = uniqueID
	entry.Favorite = favorite
	entry.UpdatedAt = now.Unix()
	entries[k] = entry
	return save(entries)
}

func save(entries map[string]Entry) error {
	list := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		list = append(list, entry)
	}
	// Evict least-recently-updated first, so accounts nobody has touched in a
	// while cannot hold the cap against ones still in use.
	if len(list) > maxEntries {
		sort.Slice(list, func(i, j int) bool { return list[i].UpdatedAt > list[j].UpdatedAt })
		list = list[:maxEntries]
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].PlatformKey != list[j].PlatformKey {
			return list[i].PlatformKey < list[j].PlatformKey
		}
		return list[i].UniqueID < list[j].UniqueID
	})

	raw, err := json.MarshalIndent(storeFile{Version: Version, Entries: list}, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(path, raw, 0o600)
}
