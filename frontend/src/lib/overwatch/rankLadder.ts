import { Role } from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/models.js";

export { Role };

/** Fixed display order for the four role columns. */
export const ROLES: Role[] = [Role.RoleTank, Role.RoleDamage, Role.RoleSupport, Role.RoleOpen];

export const ROLE_LABELS: Record<Role, string> = {
  [Role.$zero]: "",
  [Role.RoleTank]: "Tank",
  [Role.RoleDamage]: "Damage",
  [Role.RoleSupport]: "Support",
  [Role.RoleOpen]: "Open Queue",
};

/**
 * Ladder order, lowest first. Mirrors internal/owrank.Tiers exactly - Emerald
 * sits between Platinum and Diamond, and there is no Top 500 tier since it is
 * a leaderboard placement, not a rank.
 */
export const TIERS = [
  "bronze",
  "silver",
  "gold",
  "platinum",
  "emerald",
  "diamond",
  "master",
  "grandmaster",
  "champion",
] as const;
export type Tier = (typeof TIERS)[number];

export const MIN_DIVISION = 1;
export const MAX_DIVISION = 5;

export function tierIndex(tier: string): number {
  return TIERS.indexOf(tier as Tier);
}

export function tierLabel(tier: string): string {
  if (!tier) return "";
  return tier.charAt(0).toUpperCase() + tier.slice(1);
}

/**
 * Orders ranks for sorting: higher is better, unranked (empty tier) sorts
 * lowest of all. Division counts down as a player climbs (division 1 beats
 * division 5 within a tier), so it is inverted before stacking on top of the
 * tier's block - mirrors internal/owrank.RoleRank.Score exactly.
 */
export function rankScore(tier: string, division: number): number {
  const idx = tierIndex(tier);
  if (idx < 0) return -1;
  const div = division < MIN_DIVISION || division > MAX_DIVISION ? MAX_DIVISION : division;
  return idx * MAX_DIVISION + (MAX_DIVISION - div);
}
