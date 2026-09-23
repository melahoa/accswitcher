package owrank

import (
	"testing"
	"time"

	"TcNo-Acc-Switcher/internal/paths"
)

const (
	testPlatform = "BattleNet"
	testID       = "Player#1234"
)

var testNow = time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)

func useStoreRoot(t *testing.T) {
	t.Helper()
	paths.ResetForTest(t.TempDir())
}

func sampleRoles() map[Role]RoleRank {
	return map[Role]RoleRank{
		RoleTank:    {Tier: "diamond", Division: 3},
		RoleDamage:  {Tier: "gold", Division: 1},
		RoleSupport: {Tier: "champion", Division: 5},
	}
}

func TestLoadReturnsEmptyWhenTheFileIsAbsent(t *testing.T) {
	useStoreRoot(t)
	got, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Load = %#v, want empty", got)
	}
}

func TestGetReturnsEmptyRolesForAnUnknownAccount(t *testing.T) {
	useStoreRoot(t)
	entry, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(entry.Roles) != 0 {
		t.Fatalf("Roles = %#v, want empty", entry.Roles)
	}
}

func TestPutAndGetRoundTrip(t *testing.T) {
	useStoreRoot(t)
	if err := Put(testPlatform, testID, sampleRoles(), testNow); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(got.Roles) != 3 {
		t.Fatalf("Roles = %#v, want 3 entries", got.Roles)
	}
	if got.Roles[RoleTank] != (RoleRank{Tier: "diamond", Division: 3}) {
		t.Fatalf("tank rank = %#v", got.Roles[RoleTank])
	}
	if _, open := got.Roles[RoleOpen]; open {
		t.Fatalf("open should be absent (never set), got %#v", got.Roles[RoleOpen])
	}
}

func TestUnrankedRoleIsDroppedRatherThanStored(t *testing.T) {
	useStoreRoot(t)
	roles := sampleRoles()
	roles[RoleOpen] = RoleRank{} // explicit but unranked
	if err := Put(testPlatform, testID, roles, testNow); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, ok := got.Roles[RoleOpen]; ok {
		t.Fatalf("open should have been dropped, got %#v", got.Roles[RoleOpen])
	}
}

func TestClearingARoleLeavesTheOthersAlone(t *testing.T) {
	useStoreRoot(t)
	if err := Put(testPlatform, testID, sampleRoles(), testNow); err != nil {
		t.Fatalf("Put: %v", err)
	}
	next := sampleRoles()
	delete(next, RoleDamage)
	if err := Put(testPlatform, testID, next, testNow); err != nil {
		t.Fatalf("Put (clear): %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if _, ok := got.Roles[RoleDamage]; ok {
		t.Fatalf("damage should be cleared, got %#v", got.Roles[RoleDamage])
	}
	if got.Roles[RoleTank] != (RoleRank{Tier: "diamond", Division: 3}) {
		t.Fatalf("tank should be unaffected, got %#v", got.Roles[RoleTank])
	}
}

func TestPutRejectsInvalidTierOrDivision(t *testing.T) {
	useStoreRoot(t)
	bad := map[Role]RoleRank{RoleTank: {Tier: "unobtanium", Division: 3}}
	if err := Put(testPlatform, testID, bad, testNow); err == nil {
		t.Fatal("expected an error for an invalid tier")
	}
	bad = map[Role]RoleRank{RoleTank: {Tier: "gold", Division: 9}}
	if err := Put(testPlatform, testID, bad, testNow); err == nil {
		t.Fatal("expected an error for an out-of-range division")
	}
}

func TestScoreOrdersHigherTierAboveAnyDivisionOfTheOneBelow(t *testing.T) {
	bronze1 := RoleRank{Tier: "bronze", Division: 1}
	silver5 := RoleRank{Tier: "silver", Division: 5}
	if silver5.Score() <= bronze1.Score() {
		t.Fatalf("Silver 5 (%d) should outscore Bronze 1 (%d)", silver5.Score(), bronze1.Score())
	}
}

func TestScoreOrdersLowerDivisionNumberAboveHigherWithinATier(t *testing.T) {
	gold5 := RoleRank{Tier: "gold", Division: 5}
	gold1 := RoleRank{Tier: "gold", Division: 1}
	if gold1.Score() <= gold5.Score() {
		t.Fatalf("Gold 1 (%d) should outscore Gold 5 (%d)", gold1.Score(), gold5.Score())
	}
}

func TestUnrankedScoresBelowEveryRankedTier(t *testing.T) {
	unranked := RoleRank{}
	bronze5 := RoleRank{Tier: "bronze", Division: 5}
	if unranked.Score() >= bronze5.Score() {
		t.Fatalf("unranked (%d) should score below Bronze 5 (%d)", unranked.Score(), bronze5.Score())
	}
}

func TestSetHiddenAndGetRoundTrip(t *testing.T) {
	useStoreRoot(t)
	if err := SetHidden(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetHidden: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Hidden {
		t.Fatal("expected Hidden = true")
	}
}

func TestPutPreservesHiddenFlag(t *testing.T) {
	useStoreRoot(t)
	if err := SetHidden(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetHidden: %v", err)
	}
	if err := Put(testPlatform, testID, sampleRoles(), testNow); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Hidden {
		t.Fatal("Put should not have cleared Hidden")
	}
	if len(got.Roles) != 3 {
		t.Fatalf("Roles = %#v, want 3 entries", got.Roles)
	}
}

func TestSetHiddenPreservesExistingRoles(t *testing.T) {
	useStoreRoot(t)
	if err := Put(testPlatform, testID, sampleRoles(), testNow); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := SetHidden(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetHidden: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Hidden {
		t.Fatal("expected Hidden = true")
	}
	if len(got.Roles) != 3 {
		t.Fatalf("SetHidden should not have cleared Roles, got %#v", got.Roles)
	}
}

func TestSetFavoriteAndGetRoundTrip(t *testing.T) {
	useStoreRoot(t)
	if err := SetFavorite(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetFavorite: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Favorite {
		t.Fatal("expected Favorite = true")
	}
}

func TestPutPreservesFavoriteFlag(t *testing.T) {
	useStoreRoot(t)
	if err := SetFavorite(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetFavorite: %v", err)
	}
	if err := Put(testPlatform, testID, sampleRoles(), testNow); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Favorite {
		t.Fatal("Put should not have cleared Favorite")
	}
	if len(got.Roles) != 3 {
		t.Fatalf("Roles = %#v, want 3 entries", got.Roles)
	}
}

func TestSetFavoriteAndSetHiddenAreIndependent(t *testing.T) {
	useStoreRoot(t)
	if err := SetFavorite(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetFavorite: %v", err)
	}
	if err := SetHidden(testPlatform, testID, true, testNow); err != nil {
		t.Fatalf("SetHidden: %v", err)
	}
	got, err := Get(testPlatform, testID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.Favorite {
		t.Fatal("SetHidden should not have cleared Favorite")
	}
	if !got.Hidden {
		t.Fatal("expected Hidden = true")
	}
}

func TestEmeraldSitsBetweenPlatinumAndDiamond(t *testing.T) {
	platIdx := TierIndex("platinum")
	emeraldIdx := TierIndex("emerald")
	diamondIdx := TierIndex("diamond")
	if !(platIdx < emeraldIdx && emeraldIdx < diamondIdx) {
		t.Fatalf("expected platinum < emerald < diamond, got %d, %d, %d", platIdx, emeraldIdx, diamondIdx)
	}
}
