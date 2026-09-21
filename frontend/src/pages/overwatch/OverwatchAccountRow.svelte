<script lang="ts">
  import AccountLiveSessionIndicator from "../../components/AccountLiveSessionIndicator.svelte";
  import { ROLES, ROLE_LABELS, tierLabel } from "../../lib/overwatch/rankLadder";
  import { ROLE_ICONS, TIER_ICONS } from "../../lib/overwatch/icons";
  import { platformIconSrc } from "../../lib/platformIcon";
  import { PLATFORM_NAME } from "../../lib/overwatch/icons";
  import type { OverwatchAccountRowData } from "./overwatchTypes";

  export let row: OverwatchAccountRowData;
  export let onLogin: () => void;
  export let onEdit: () => void;

  function handleRowClick(e: MouseEvent): void {
    // The edit button is inside the row; let its own click handler run instead
    // of also logging in.
    if ((e.target as HTMLElement).closest(".ow-edit-btn")) return;
    onLogin();
  }

  function handleRowKeydown(e: KeyboardEvent): void {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onLogin();
    }
  }
</script>

<div
  class="ow-row"
  class:ow-row--active={row.currentSession}
  role="button"
  tabindex="0"
  on:click={handleRowClick}
  on:keydown={handleRowKeydown}
>
  <AccountLiveSessionIndicator active={row.currentSession} tooltipText="Currently signed in" />

  <div class="ow-avatar">
    {#if row.imageUrl}
      <img src={row.imageUrl} alt="" />
    {:else}
      <div class="ow-avatar-fallback">{row.name.charAt(0).toUpperCase()}</div>
    {/if}
    <img class="ow-platform-badge" src={platformIconSrc(PLATFORM_NAME[row.platform])} alt={PLATFORM_NAME[row.platform]} />
  </div>

  <div class="ow-name" title={row.name}>{row.name}</div>

  <div class="ow-roles">
    {#each ROLES as role (role)}
      {@const rank = row.roles[role]}
      <div class="ow-role-badge" class:ow-role-badge--unranked={!rank} title="{ROLE_LABELS[role]}: {rank ? `${tierLabel(rank.tier)} ${rank.division}` : 'Unranked'}">
        <img class="ow-role-icon" src={ROLE_ICONS[role]} alt={ROLE_LABELS[role]} />
        {#if rank}
          <img class="ow-tier-icon" src={TIER_ICONS[rank.tier]} alt={tierLabel(rank.tier)} />
          <span class="ow-division">{rank.division}</span>
        {:else}
          <span class="ow-division ow-division--unranked">-</span>
        {/if}
      </div>
    {/each}
  </div>

  <button type="button" class="ow-edit-btn" on:click|stopPropagation={onEdit} aria-label="Edit ranks for {row.name}">
    Edit
  </button>
</div>

<style>
  .ow-row {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.6rem 0.85rem;
    border-radius: 8px;
    background: var(--overlay-white-06, rgba(255, 255, 255, 0.06));
    cursor: pointer;
    outline: none;
  }
  .ow-row:hover {
    background: var(--overlay-white-08, rgba(255, 255, 255, 0.08));
  }
  .ow-row:focus-visible {
    box-shadow: 0 0 0 2px var(--accent);
  }
  .ow-row--active {
    background: var(--overlay-white-12, rgba(255, 255, 255, 0.12));
  }

  .ow-avatar {
    position: relative;
    flex: 0 0 auto;
    width: 40px;
    height: 40px;
  }
  .ow-avatar img,
  .ow-avatar-fallback {
    width: 40px;
    height: 40px;
    border-radius: 6px;
    object-fit: cover;
  }
  .ow-avatar-fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--code, #263440);
    color: var(--whiteSecondary, #fff);
    font-weight: 700;
  }
  .ow-platform-badge {
    position: absolute;
    right: -4px;
    bottom: -4px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    box-shadow: 0 0 0 2px var(--mainContentBackground, #14181f);
  }

  .ow-name {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--whiteSecondary, #fff);
    font-weight: 600;
  }

  .ow-roles {
    display: flex;
    gap: 0.5rem;
    flex: 0 0 auto;
  }
  .ow-role-badge {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.2rem 0.4rem;
    border-radius: 6px;
    background: var(--overlay-white-06, rgba(255, 255, 255, 0.06));
    min-width: 3.4rem;
  }
  .ow-role-badge--unranked {
    opacity: 0.45;
  }
  .ow-role-icon {
    width: 16px;
    height: 16px;
  }
  .ow-tier-icon {
    width: 16px;
    height: 16px;
  }
  .ow-division {
    font-size: 0.8rem;
    color: var(--whiteSecondary, #fff);
  }
  .ow-division--unranked {
    color: var(--text-dim-gray, #6b6a6a);
  }

  .ow-edit-btn {
    flex: 0 0 auto;
    padding: 0.35rem 0.7rem;
    border-radius: 6px;
    border: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
    background: transparent;
    color: var(--whiteSecondary, #fff);
    font-size: 0.8rem;
    cursor: pointer;
  }
  .ow-edit-btn:hover {
    background: var(--overlay-white-08, rgba(255, 255, 255, 0.08));
    border-color: var(--accent);
  }
</style>
