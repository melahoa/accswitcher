package owrank

import (
	"time"

	"TcNo-Acc-Switcher/internal/security"
)

// RoleRankDTO is the JSON shape sent to and from the frontend for one role.
type RoleRankDTO struct {
	Tier     string `json:"tier"`
	Division int    `json:"division"`
}

// EntryDTO is the JSON shape of one account's full set of ranks.
type EntryDTO struct {
	PlatformKey string               `json:"platformKey"`
	UniqueID    string               `json:"uniqueId"`
	Roles       map[Role]RoleRankDTO `json:"roles"`
	Hidden      bool                 `json:"hidden"`
}

func toDTO(e Entry) EntryDTO {
	roles := make(map[Role]RoleRankDTO, len(e.Roles))
	for role, rr := range e.Roles {
		roles[role] = RoleRankDTO{Tier: rr.Tier, Division: rr.Division}
	}
	return EntryDTO{PlatformKey: e.PlatformKey, UniqueID: e.UniqueID, Roles: roles, Hidden: e.Hidden}
}

// Service is the Wails-bound entry point the Overwatch build's frontend calls
// to read and write manually-entered ranks. It holds no state of its own -
// everything lives in the on-disk store - so it needs no constructor.
type Service struct{}

// ServiceName is the Wails service name.
func (s *Service) ServiceName() string { return "OverwatchService" }

// GetRank returns one account's ranks across all four roles. An account with
// no ranks set yet comes back with an empty Roles map rather than an error.
func (s *Service) GetRank(platformKey, uniqueId string) (EntryDTO, error) {
	if err := security.RequireUnlocked(); err != nil {
		return EntryDTO{}, err
	}
	entry, err := Get(platformKey, uniqueId)
	if err != nil {
		return EntryDTO{}, err
	}
	return toDTO(entry), nil
}

// SetRanks replaces all four roles for an account at once. Omitting a role
// from the map (or sending it with an empty Tier) clears that role back to
// unranked.
func (s *Service) SetRanks(platformKey, uniqueId string, roles map[Role]RoleRankDTO) error {
	if err := security.RequireUnlocked(); err != nil {
		return err
	}
	converted := make(map[Role]RoleRank, len(roles))
	for role, rr := range roles {
		converted[role] = RoleRank{Tier: rr.Tier, Division: rr.Division}
	}
	return Put(platformKey, uniqueId, converted, time.Now())
}

// SetHidden shows or hides an account in the Overwatch build's own account
// list. It never touches the account's real login files, so the account
// still works normally everywhere else (including the main app).
func (s *Service) SetHidden(platformKey, uniqueId string, hidden bool) error {
	if err := security.RequireUnlocked(); err != nil {
		return err
	}
	return SetHidden(platformKey, uniqueId, hidden, time.Now())
}
