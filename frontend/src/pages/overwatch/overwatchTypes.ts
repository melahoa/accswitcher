import type { Role } from "../../lib/overwatch/rankLadder";
import type { OverwatchPlatform } from "../../lib/overwatch/icons";

export type OverwatchRoleRank = { tier: string; division: number };

export type OverwatchRoleRanks = Partial<Record<Role, OverwatchRoleRank>>;

export interface OverwatchAccountRowData {
  platform: OverwatchPlatform;
  /** The exact platform key the backend expects ("Steam" or "BattleNet"). */
  platformKey: string;
  id: string;
  name: string;
  imageUrl: string;
  currentSession: boolean;
  roles: OverwatchRoleRanks;
}
