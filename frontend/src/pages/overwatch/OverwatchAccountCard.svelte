<script lang="ts">
  import AccountLiveSessionIndicator from "../../components/AccountLiveSessionIndicator.svelte";
  import { ROLES, ROLE_LABELS, tierLabel } from "../../lib/overwatch/rankLadder";
  import { ROLE_ICONS, TIER_ICONS, PLATFORM_ICONS, PLATFORM_LABEL } from "../../lib/overwatch/icons";
  import type { OverwatchAccountRowData } from "./overwatchTypes";

  export let row: OverwatchAccountRowData;
  export let onLogin: () => void;
  export let onEdit: () => void;

  function handleCardClick(e: MouseEvent): void {
    // The edit button sits inside the card; let its own handler run instead of
    // also logging in.
    if ((e.target as HTMLElement).closest(".ow-card-edit")) return;
    onLogin();
  }

  function handleCardKeydown(e: KeyboardEvent): void {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      onLogin();
    }
  }
</script>

<div
  class="ow-card"
  class:ow-card--active={row.currentSession}
  class:ow-card--hidden={row.hidden}
  role="button"
  tabindex="0"
  on:click={handleCardClick}
  on:keydown={handleCardKeydown}
>
  <img
    class="ow-card-platform-badge"
    src={PLATFORM_ICONS[row.platform]}
    alt={PLATFORM_LABEL[row.platform]}
    title={PLATFORM_LABEL[row.platform]}
  />

  <button type="button" class="ow-card-edit" on:click|stopPropagation={onEdit} aria-label="Edit ranks for {row.name}">
    <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true">
      <path
        fill="currentColor"
        d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25ZM20.71 7.04a1 1 0 0 0 0-1.41l-2.34-2.34a1 1 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83Z"
      />
    </svg>
  </button>

  <div class="ow-card-avatar">
    {#if row.imageUrl}
      <img src={row.imageUrl} alt="" />
    {:else}
      <div class="ow-card-avatar-fallback">{row.name.charAt(0).toUpperCase()}</div>
    {/if}
    <AccountLiveSessionIndicator active={row.currentSession} tooltipText="Currently signed in" />
  </div>

  <div class="ow-card-name" title={row.name}>{row.name}</div>
  {#if row.accountName && row.accountName !== row.name}
    <div class="ow-card-username" title={row.accountName}>@{row.accountName}</div>
  {/if}

  <div class="ow-card-roles">
    {#each ROLES as role (role)}
      {@const rank = row.roles[role]}
      <div
        class="ow-card-role"
        class:ow-card-role--unranked={!rank}
        title="{ROLE_LABELS[role]}: {rank ? `${tierLabel(rank.tier)} ${rank.division}` : 'Unranked'}"
      >
        <img class="ow-card-role-icon" src={ROLE_ICONS[role]} alt={ROLE_LABELS[role]} />
        {#if rank}
          <img class="ow-card-tier-icon" src={TIER_ICONS[rank.tier]} alt={tierLabel(rank.tier)} />
          <span class="ow-card-division">{rank.division}</span>
        {:else}
          <span class="ow-card-division ow-card-division--unranked">&ndash;</span>
        {/if}
      </div>
    {/each}
  </div>
</div>

<style>
  .ow-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem 0.85rem 0.85rem;
    border-radius: 14px;
    background: var(--overlay-white-06, rgba(255, 255, 255, 0.06));
    border: 1px solid var(--overlay-white-08, rgba(255, 255, 255, 0.08));
    cursor: pointer;
    outline: none;
    transition: transform 120ms ease, border-color 120ms ease, background 120ms ease;
  }
  .ow-card:hover {
    background: var(--overlay-white-08, rgba(255, 255, 255, 0.08));
    border-color: var(--accent);
    transform: translateY(-2px);
  }
  .ow-card:focus-visible {
    box-shadow: 0 0 0 2px var(--accent);
  }
  .ow-card--active {
    background: var(--overlay-white-12, rgba(255, 255, 255, 0.12));
    border-color: color-mix(in srgb, var(--accent) 50%, transparent);
  }
  .ow-card--hidden {
    opacity: 0.55;
  }
  .ow-card--hidden:hover {
    opacity: 0.85;
  }

  .ow-card-edit {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    border-radius: 50%;
    border: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
    background: var(--mainContentBackground, #14181f);
    color: var(--whiteSecondary, #fff);
    cursor: pointer;
    opacity: 0;
    transition: opacity 120ms ease, border-color 120ms ease;
  }
  .ow-card:hover .ow-card-edit,
  .ow-card:focus-within .ow-card-edit {
    opacity: 1;
  }
  .ow-card-edit:hover {
    border-color: var(--accent);
    color: var(--accent);
  }

  .ow-card-avatar {
    position: relative;
    width: 84px;
    height: 84px;
  }
  .ow-card-avatar img,
  .ow-card-avatar-fallback {
    width: 84px;
    height: 84px;
    border-radius: 12px;
    object-fit: cover;
  }
  .ow-card-avatar-fallback {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--code, #263440);
    color: var(--whiteSecondary, #fff);
    font-size: 1.8rem;
    font-weight: 700;
  }
  .ow-card-platform-badge {
    position: absolute;
    top: 0.5rem;
    left: 0.5rem;
    z-index: 2;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    box-shadow: 0 0 0 3px var(--mainContentBackground, #14181f);
  }

  .ow-card-name {
    width: 100%;
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--whiteSecondary, #fff);
    font-weight: 600;
    font-size: 1rem;
  }
  .ow-card-username {
    width: 100%;
    margin-top: -0.3rem;
    text-align: center;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text-body-muted, #9d9d9d);
    font-size: 0.78rem;
  }

  .ow-card-roles {
    /* Pushed to the bottom of the card via the flex auto-margin trick, so it
       lands in the same place whether or not the username line above it is
       present - grid stretches every card in a row to the tallest one's
       height, and without this the roles block would float at a different
       height per card instead of sharing one baseline. */
    margin-top: auto;
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
    width: 100%;
  }
  .ow-card-role {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    padding: 0.5rem 0.4rem;
    border-radius: 8px;
    background: var(--overlay-white-06, rgba(255, 255, 255, 0.06));
  }
  .ow-card-role--unranked {
    opacity: 0.4;
  }
  .ow-card-role-icon {
    width: 26px;
    height: 26px;
    flex: 0 0 auto;
  }
  .ow-card-tier-icon {
    width: 30px;
    height: 30px;
    flex: 0 0 auto;
  }
  .ow-card-division {
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--whiteSecondary, #fff);
  }
  .ow-card-division--unranked {
    color: var(--text-dim-gray, #6b6a6a);
  }
</style>
