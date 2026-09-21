<script lang="ts">
  import { onMount } from "svelte";
  import * as BasicService from "../../../bindings/TcNo-Acc-Switcher/internal/basic/basicservice.js";
  import * as SteamService from "../../../bindings/TcNo-Acc-Switcher/internal/steam/steamservice.js";
  import * as OverwatchService from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/service.js";
  import type { RoleRankDTO } from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/models.js";
  import { pushToast } from "../../stores/toast";
  import { formatToastWithError } from "../../lib/formatWailsError";
  import { ROLES, Role, ROLE_LABELS, rankScore } from "../../lib/overwatch/rankLadder";
  import OverwatchAccountRow from "./OverwatchAccountRow.svelte";
  import OverwatchRankEditor from "./OverwatchRankEditor.svelte";
  import type { OverwatchAccountRowData, OverwatchRoleRanks } from "./overwatchTypes";

  type SortKey = Role | "name";

  let rows: OverwatchAccountRowData[] = [];
  let loading = true;
  let loadError = "";
  let sortBy: SortKey = "name";
  let editingRow: OverwatchAccountRowData | null = null;

  function rowKey(row: { platform: string; id: string }): string {
    return `${row.platform}|${row.id}`;
  }

  async function loadRolesFor(platformKey: string, uniqueId: string): Promise<OverwatchRoleRanks> {
    try {
      const entry = await OverwatchService.GetRank(platformKey, uniqueId);
      const roles: OverwatchRoleRanks = {};
      for (const role of ROLES) {
        const rr = entry.roles[role];
        if (rr && rr.tier) roles[role] = { tier: rr.tier, division: rr.division };
      }
      return roles;
    } catch {
      // An account with no ranks saved yet, or a store read failure - either
      // way the row just shows every role as unranked.
      return {};
    }
  }

  async function loadAll(): Promise<void> {
    loading = true;
    loadError = "";
    try {
      const [battlenetAccounts, steamAccounts] = await Promise.all([
        BasicService.GetAccounts("BattleNet").catch(() => []),
        SteamService.GetSteamAccounts().catch(() => []),
      ]);

      const battlenetRows = await Promise.all(
        battlenetAccounts.map(
          async (a): Promise<OverwatchAccountRowData> => ({
            platform: "battlenet",
            platformKey: "BattleNet",
            id: a.uniqueId,
            name: a.displayName,
            imageUrl: a.imageUrl,
            currentSession: a.currentSession,
            roles: await loadRolesFor("BattleNet", a.uniqueId),
          }),
        ),
      );
      const steamRows = await Promise.all(
        steamAccounts.map(
          async (a): Promise<OverwatchAccountRowData> => ({
            platform: "steam",
            platformKey: "Steam",
            id: a.steamId64,
            name: a.displayName || a.personaName,
            imageUrl: a.imageUrl,
            currentSession: a.currentSession,
            roles: await loadRolesFor("Steam", a.steamId64),
          }),
        ),
      );

      rows = [...battlenetRows, ...steamRows];
    } catch (err) {
      loadError = formatToastWithError("Could not load accounts", err);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void loadAll();
  });

  async function handleLogin(row: OverwatchAccountRowData): Promise<void> {
    try {
      if (row.platform === "battlenet") {
        await BasicService.SwapToAccount(row.platformKey, row.id, []);
      } else {
        // personaState -1 keeps whatever it already was - same call PlatformSteam.svelte makes.
        await SteamService.SwapToSteamAccount(row.id, -1, []);
      }
    } catch (err) {
      pushToast({ type: "error", message: formatToastWithError(`Could not sign in to ${row.name}`, err) });
    }
  }

  function openEditor(row: OverwatchAccountRowData): void {
    editingRow = row;
  }

  function closeEditor(): void {
    editingRow = null;
  }

  async function handleSaveRanks(roles: OverwatchRoleRanks): Promise<void> {
    const target = editingRow;
    if (!target) return;
    const payload: Partial<Record<Role, RoleRankDTO>> = {};
    for (const role of ROLES) {
      const rr = roles[role];
      payload[role] = { tier: rr?.tier ?? "", division: rr?.division ?? 0 };
    }
    try {
      await OverwatchService.SetRanks(target.platformKey, target.id, payload);
      rows = rows.map((r) => (rowKey(r) === rowKey(target) ? { ...r, roles } : r));
      editingRow = null;
    } catch (err) {
      pushToast({ type: "error", message: formatToastWithError("Could not save rank", err) });
    }
  }

  $: sortedRows = [...rows].sort((a, b) => {
    if (sortBy === "name") return a.name.localeCompare(b.name);
    const rankA = a.roles[sortBy];
    const rankB = b.roles[sortBy];
    const scoreA = rankA ? rankScore(rankA.tier, rankA.division) : -1;
    const scoreB = rankB ? rankScore(rankB.tier, rankB.division) : -1;
    if (scoreA !== scoreB) return scoreB - scoreA;
    return a.name.localeCompare(b.name);
  });

  const sortOptions: { key: SortKey; label: string }[] = [
    { key: "name", label: "Name" },
    ...ROLES.map((role) => ({ key: role, label: ROLE_LABELS[role] })),
  ];
</script>

<div class="ow-list-page">
  <div class="ow-toolbar">
    <span class="ow-toolbar-label">Sort by</span>
    <div class="ow-sort-group" role="group" aria-label="Sort accounts by">
      {#each sortOptions as opt (opt.key)}
        <button
          type="button"
          class="ow-sort-btn"
          class:ow-sort-btn--active={sortBy === opt.key}
          on:click={() => (sortBy = opt.key)}
        >
          {opt.label}
        </button>
      {/each}
    </div>
  </div>

  <div class="ow-list">
    {#if loading}
      <p class="ow-status">Loading accounts...</p>
    {:else if loadError}
      <p class="ow-status ow-status--error">{loadError}</p>
    {:else if sortedRows.length === 0}
      <p class="ow-status">No Battle.net or Steam accounts found yet. Sign in once from within each app and it will show up here.</p>
    {:else}
      {#each sortedRows as row (rowKey(row))}
        <OverwatchAccountRow {row} onLogin={() => handleLogin(row)} onEdit={() => openEditor(row)} />
      {/each}
    {/if}
  </div>
</div>

{#if editingRow}
  <OverwatchRankEditor row={editingRow} onSave={handleSaveRanks} onCancel={closeEditor} />
{/if}

<style>
  .ow-list-page {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    padding: 1rem;
    gap: 0.75rem;
    overflow: hidden;
  }
  .ow-toolbar {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex: 0 0 auto;
  }
  .ow-toolbar-label {
    font-size: 0.8rem;
    color: var(--text-body-muted, #9d9d9d);
  }
  .ow-sort-group {
    display: flex;
    gap: 0.35rem;
    flex-wrap: wrap;
  }
  .ow-sort-btn {
    padding: 0.3rem 0.65rem;
    border-radius: 999px;
    border: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
    background: transparent;
    color: var(--whiteSecondary, #fff);
    font-size: 0.78rem;
    cursor: pointer;
  }
  .ow-sort-btn:hover {
    background: var(--overlay-white-08, rgba(255, 255, 255, 0.08));
  }
  .ow-sort-btn--active {
    background: var(--accent);
    border-color: var(--accent);
    color: #0b0e12;
    font-weight: 600;
  }

  .ow-list {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }
  .ow-status {
    color: var(--text-body-muted, #9d9d9d);
    font-size: 0.85rem;
  }
  .ow-status--error {
    color: var(--error-text-soft, #ffb4b4);
    white-space: pre-line;
  }
</style>
