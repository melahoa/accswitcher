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
  /** Steam's login username, distinct from its public persona name. Empty for Battle.net. */
  accountName: string;
  imageUrl: string;
  currentSession: boolean;
  roles: OverwatchRoleRanks;
  /** App-only visibility - never touches the real account or its login files. */
  hidden: boolean;
  /** Pins this account above every non-favorite account in the list. */
  favorite: boolean;
  note: string;
}
