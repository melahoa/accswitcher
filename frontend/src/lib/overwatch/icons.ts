// Every icon here is a placeholder (see frontend/src/assets/overwatch/) - swap
// the PNG files in place and nothing here needs to change.
import { Role } from "./rankLadder";

import tankIcon from "../../assets/overwatch/roles/tank.png";
import damageIcon from "../../assets/overwatch/roles/damage.png";
import supportIcon from "../../assets/overwatch/roles/support.png";
import openIcon from "../../assets/overwatch/roles/open.png";

export const ROLE_ICONS: Record<Role, string> = {
  [Role.$zero]: "",
  [Role.RoleTank]: tankIcon,
  [Role.RoleDamage]: damageIcon,
  [Role.RoleSupport]: supportIcon,
  [Role.RoleOpen]: openIcon,
};

import bronzeIcon from "../../assets/overwatch/ranks/bronze.png";
import silverIcon from "../../assets/overwatch/ranks/silver.png";
import goldIcon from "../../assets/overwatch/ranks/gold.png";
import platinumIcon from "../../assets/overwatch/ranks/platinum.png";
import emeraldIcon from "../../assets/overwatch/ranks/emerald.png";
import diamondIcon from "../../assets/overwatch/ranks/diamond.png";
import masterIcon from "../../assets/overwatch/ranks/master.png";
import grandmasterIcon from "../../assets/overwatch/ranks/grandmaster.png";
import championIcon from "../../assets/overwatch/ranks/champion.png";

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

import steamIcon from "../../assets/overwatch/platforms/steam.png";
import battlenetIcon from "../../assets/overwatch/platforms/battlenet.png";

export type OverwatchPlatform = "steam" | "battlenet";

export const PLATFORM_ICONS: Record<OverwatchPlatform, string> = {
  steam: steamIcon,
  battlenet: battlenetIcon,
};

export const PLATFORM_LABEL: Record<OverwatchPlatform, string> = {
  steam: "Steam",
  battlenet: "Battle.net",
};
