// Every icon here is a placeholder (see frontend/src/assets/overwatch/) - swap
// the SVG files in place and nothing here needs to change.
import { Role } from "./rankLadder";

import tankIcon from "../../assets/overwatch/roles/tank.svg";
import damageIcon from "../../assets/overwatch/roles/damage.svg";
import supportIcon from "../../assets/overwatch/roles/support.svg";
import openIcon from "../../assets/overwatch/roles/open.svg";

export const ROLE_ICONS: Record<Role, string> = {
  [Role.$zero]: "",
  [Role.RoleTank]: tankIcon,
  [Role.RoleDamage]: damageIcon,
  [Role.RoleSupport]: supportIcon,
  [Role.RoleOpen]: openIcon,
};

import bronzeIcon from "../../assets/overwatch/ranks/bronze.svg";
import silverIcon from "../../assets/overwatch/ranks/silver.svg";
import goldIcon from "../../assets/overwatch/ranks/gold.svg";
import platinumIcon from "../../assets/overwatch/ranks/platinum.svg";
import emeraldIcon from "../../assets/overwatch/ranks/emerald.svg";
import diamondIcon from "../../assets/overwatch/ranks/diamond.svg";
import masterIcon from "../../assets/overwatch/ranks/master.svg";
import grandmasterIcon from "../../assets/overwatch/ranks/grandmaster.svg";
import championIcon from "../../assets/overwatch/ranks/champion.svg";

export const TIER_ICONS: Record<string, string> = {
  bronze: bronzeIcon,
  silver: silverIcon,
  gold: goldIcon,
  platinum: platinumIcon,
  emerald: emeraldIcon,
  diamond: diamondIcon,
  master: masterIcon,
  grandmaster: grandmasterIcon,
  champion: championIcon,
};

export type OverwatchPlatform = "steam" | "battlenet";

/** Maps to the platform art already bundled for every build variant. */
export const PLATFORM_NAME: Record<OverwatchPlatform, string> = {
  steam: "Steam",
  battlenet: "BattleNet",
};
