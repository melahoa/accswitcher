<script lang="ts">
  import { ROLES, ROLE_LABELS, TIERS, tierLabel, MIN_DIVISION, MAX_DIVISION } from "../../lib/overwatch/rankLadder";
  import { ROLE_ICONS } from "../../lib/overwatch/icons";
  import type { OverwatchAccountRowData, OverwatchRoleRanks } from "./overwatchTypes";

  export let row: OverwatchAccountRowData;
  export let onSave: (roles: OverwatchRoleRanks) => void;
  export let onCancel: () => void;

  const divisions = Array.from({ length: MAX_DIVISION - MIN_DIVISION + 1 }, (_, i) => MIN_DIVISION + i);

  // Local working copy - the account's stored ranks are not touched until Save.
  let draft: Record<string, { tier: string; division: number }> = {};
  for (const role of ROLES) {
    const existing = row.roles[role];
    draft[role] = { tier: existing?.tier ?? "", division: existing?.division ?? MAX_DIVISION };
  }

  function setTier(role: string, tier: string): void {
    draft[role] = { ...draft[role], tier };
  }

  function setDivision(role: string, division: number): void {
    draft[role] = { ...draft[role], division };
  }

  function handleSave(): void {
    const roles: OverwatchRoleRanks = {};
    for (const role of ROLES) {
      const d = draft[role];
      if (d.tier) {
        roles[role] = { tier: d.tier, division: d.division };
      }
    }
    onSave(roles);
  }

  function handleBackdropKeydown(e: KeyboardEvent): void {
    if (e.key === "Escape") onCancel();
  }
</script>

<svelte:window on:keydown={handleBackdropKeydown} />

<div
  class="ow-editor-backdrop"
  role="presentation"
  on:click={(e) => { if (e.target === e.currentTarget) onCancel(); }}
>
  <div
    class="ow-editor"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-label="Edit ranks for {row.name}"
  >
    <h2 class="ow-editor-title">{row.name}</h2>

    <div class="ow-editor-roles">
      {#each ROLES as role (role)}
        <div class="ow-editor-role">
          <img class="ow-editor-role-icon" src={ROLE_ICONS[role]} alt="" />
          <span class="ow-editor-role-label">{ROLE_LABELS[role]}</span>
          <select
            class="ow-editor-select"
            value={draft[role].tier}
            on:change={(e) => setTier(role, (e.target as HTMLSelectElement).value)}
          >
            <option value="">Unranked</option>
            {#each TIERS as tier (tier)}
              <option value={tier}>{tierLabel(tier)}</option>
            {/each}
          </select>
          <select
            class="ow-editor-select ow-editor-select--division"
            value={draft[role].division}
            disabled={!draft[role].tier}
            on:change={(e) => setDivision(role, Number((e.target as HTMLSelectElement).value))}
          >
            {#each divisions as division (division)}
              <option value={division}>{division}</option>
            {/each}
          </select>
        </div>
      {/each}
    </div>

    <div class="ow-editor-actions">
      <button type="button" class="ow-btn ow-btn--secondary" on:click={onCancel}>Cancel</button>
      <button type="button" class="ow-btn ow-btn--primary" on:click={handleSave}>Save</button>
    </div>
  </div>
</div>

<style>
  .ow-editor-backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--backdrop-scrim-55, rgba(0, 0, 0, 0.55));
  }
  .ow-editor {
    width: min(420px, calc(100vw - 2rem));
    max-height: calc(100vh - 4rem);
    overflow-y: auto;
    padding: 1.25rem;
    border-radius: 10px;
    background: var(--mainContentBackground, #14181f);
    box-shadow: var(--shadow, 0 4px 10px -6px #010101);
    color: var(--whiteSecondary, #fff);
  }
  .ow-editor-title {
    margin: 0 0 1rem;
    font-size: 1rem;
  }
  .ow-editor-roles {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }
  .ow-editor-role {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .ow-editor-role-icon {
    width: 18px;
    height: 18px;
    flex: 0 0 auto;
  }
  .ow-editor-role-label {
    flex: 1 1 auto;
    font-size: 0.85rem;
  }
  .ow-editor-select {
    padding: 0.3rem 0.4rem;
    border-radius: 6px;
    background: var(--code-background, #131a20);
    color: var(--whiteSecondary, #fff);
    border: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
  }
  .ow-editor-select--division {
    width: 3.2rem;
  }
  .ow-editor-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1.25rem;
  }
  .ow-btn {
    padding: 0.4rem 0.9rem;
    border-radius: 6px;
    font-size: 0.85rem;
    cursor: pointer;
    border: 1px solid transparent;
  }
  .ow-btn--secondary {
    background: transparent;
    border-color: var(--overlay-white-14, rgba(255, 255, 255, 0.14));
    color: var(--whiteSecondary, #fff);
  }
  .ow-btn--primary {
    background: var(--accent);
    color: #0b0e12;
    font-weight: 600;
  }
</style>
