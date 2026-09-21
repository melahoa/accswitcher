<script lang="ts">
  import { onMount } from "svelte";
  import * as BasicService from "../../../bindings/TcNo-Acc-Switcher/internal/basic/basicservice.js";
  import * as SteamService from "../../../bindings/TcNo-Acc-Switcher/internal/steam/steamservice.js";
  import type { Settings as SteamSettings } from "../../../bindings/TcNo-Acc-Switcher/internal/steam/models.js";
  import * as OverwatchService from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/service.js";
  import type { RoleRankDTO } from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/models.js";
  import { pushToast } from "../../stores/toast";
  import { formatToastWithError } from "../../lib/formatWailsError";
  import { PLATFORM_ICONS } from "../../lib/overwatch/icons";
  import { ROLES, Role, ROLE_LABELS, rankScore } from "../../lib/overwatch/rankLadder";
  import OverwatchAccountCard from "./OverwatchAccountCard.svelte";
  import OverwatchAccountEditor from "./OverwatchAccountEditor.svelte";
  import OverwatchNamePrompt from "./OverwatchNamePrompt.svelte";
  import type { OverwatchAccountRowData, OverwatchRoleRanks } from "./overwatchTypes";

  type SortKey = Role | "name";

  let rows: OverwatchAccountRowData[] = [];
  // Only the very first load shows the blocking "Loading..." state and
  // replaces the grid; a background refresh (focus, after Save/Add, the
  // Refresh button) updates `rows` in place so the keyed each-block patches
  // existing cards instead of unmounting the whole grid - which was
  // resetting scroll position and occasionally eating a click that landed on
  // a card mid-teardown.
  let initialLoading = true;
  let refreshing = false;
  let loadError = "";
  let loadSeq = 0;
  let sortBy: SortKey = "name";
  let showHidden = false;
  let editingRow: OverwatchAccountRowData | null = null;
  let editingNote = "";
  let showSaveBattleNetPrompt = false;
  let saveBattleNetSuggestedName = "";

  const DEFAULT_STEAM_FOLDER = "C:\\Program Files (x86)\\Steam";
  let showSteamFolderPrompt = false;
  let steamFolderPromptValue = DEFAULT_STEAM_FOLDER;
  let cachedSteamSettings: SteamSettings | null = null;

  function rowKey(row: { platform: string; id: string }): string {
    return `${row.platform}|${row.id}`;
  }

  async function loadMetaFor(platformKey: string, uniqueId: string): Promise<{ roles: OverwatchRoleRanks; hidden: boolean }> {
    try {
      const entry = await OverwatchService.GetRank(platformKey, uniqueId);
      const roles: OverwatchRoleRanks = {};
      for (const role of ROLES) {
        const rr = entry.roles[role];
        if (rr && rr.tier) roles[role] = { tier: rr.tier, division: rr.division };
      }
      return { roles, hidden: entry.hidden };
    } catch {
      // An account with no ranks saved yet, or a store read failure - either
      // way the row just shows every role as unranked and visible.
      return { roles: {}, hidden: false };
    }
  }

  async function loadAll(): Promise<void> {
    const seq = ++loadSeq;
    const isInitial = initialLoading;
    if (!isInitial) refreshing = true;
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
            accountName: "",
            imageUrl: a.imageUrl,
            currentSession: a.currentSession,
            ...(await loadMetaFor("BattleNet", a.uniqueId)),
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
            accountName: a.accountName,
            imageUrl: a.imageUrl,
            currentSession: a.currentSession,
            ...(await loadMetaFor("Steam", a.steamId64)),
          }),
        ),
      );

      // A newer load already landed while this one was in flight - discard
      // this result rather than clobbering fresher data with stale data.
      if (seq !== loadSeq) return;
      rows = [...battlenetRows, ...steamRows];
      if (isInitial) loadError = "";
    } catch (err) {
      if (seq !== loadSeq) return;
      const message = formatToastWithError("Could not load accounts", err);
      if (isInitial) {
        loadError = message;
      } else {
        console.error("[overwatch] background refresh failed", err);
      }
    } finally {
      if (seq === loadSeq) {
        initialLoading = false;
        refreshing = false;
      }
    }
  }

  onMount(() => {
    void loadAll();
    // Steam accounts sync from loginusers.vdf on every fetch, and a saved
    // Battle.net session only appears after the user logs in outside this
    // window - refreshing when focus actually returns after having left
    // means coming back from either shows the result without an explicit
    // click. Paired with blur so that moving or clicking within this same
    // window (which can re-fire "focus" without ever really leaving) does
    // not trigger a refresh.
    let wasBlurred = false;
    const onBlur = () => { wasBlurred = true; };
    const onFocus = () => {
      if (!wasBlurred) return;
      wasBlurred = false;
      void loadAll();
    };
    window.addEventListener("blur", onBlur);
    window.addEventListener("focus", onFocus);
    return () => {
      window.removeEventListener("blur", onBlur);
      window.removeEventListener("focus", onFocus);
    };
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

  async function handleAddBattleNetAccount(): Promise<void> {
    try {
      await BasicService.AddNew("BattleNet");
      pushToast({
        type: "info",
        message: 'Log in to the Battle.net account you want to add, then click "Save current session" below.',
        duration: 10000,
      });
    } catch (err) {
      console.error("[overwatch] AddNew(BattleNet) failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not start adding a Battle.net account", err) });
    }
  }

  async function handleAddSteamAccount(): Promise<void> {
    try {
      await SteamService.SteamAddNew();
      pushToast({
        type: "info",
        message: 'Log in to the Steam account you want to add, then click "Refresh" below.',
        duration: 10000,
      });
    } catch (err) {
      console.error("[overwatch] SteamAddNew failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not start adding a Steam account", err) });
    }
  }

  async function openSaveBattleNetPrompt(): Promise<void> {
    try {
      saveBattleNetSuggestedName = await BasicService.SuggestedSaveAccountName("BattleNet");
    } catch (err) {
      console.error("[overwatch] SuggestedSaveAccountName(BattleNet) failed", err);
      saveBattleNetSuggestedName = "";
    }
    showSaveBattleNetPrompt = true;
  }

  async function handleConfirmSaveBattleNet(name: string): Promise<void> {
    showSaveBattleNetPrompt = false;
    pushToast({ type: "info", message: `Saving "${name}" - this will briefly close Battle.net...`, duration: 6000 });
    try {
      await BasicService.SaveCurrent("BattleNet", name);
      pushToast({ type: "success", message: `Saved ${name}.`, duration: 4000 });
      await loadAll();
    } catch (err) {
      console.error("[overwatch] SaveCurrent(BattleNet) failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not save this account", err) });
    }
  }

  async function openSteamFolderPrompt(): Promise<void> {
    try {
      cachedSteamSettings = await SteamService.GetSteamSettings();
      steamFolderPromptValue = cachedSteamSettings.FolderPath?.trim() || DEFAULT_STEAM_FOLDER;
      showSteamFolderPrompt = true;
    } catch (err) {
      console.error("[overwatch] GetSteamSettings failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not load Steam settings", err) });
    }
  }

  async function handleConfirmSteamFolder(path: string): Promise<void> {
    showSteamFolderPrompt = false;
    if (!cachedSteamSettings) return;
    try {
      cachedSteamSettings.FolderPath = path === DEFAULT_STEAM_FOLDER ? "" : path;
      await SteamService.SaveSteamSettings(cachedSteamSettings);
      pushToast({ type: "success", message: "Steam folder updated.", duration: 4000 });
      await loadAll();
    } catch (err) {
      console.error("[overwatch] SaveSteamSettings failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not save the Steam folder", err) });
    }
  }

  async function openEditor(row: OverwatchAccountRowData): Promise<void> {
    editingRow = row;
    try {
      editingNote = await BasicService.GetAccountNote(row.platformKey, row.id);
    } catch (err) {
      console.error("[overwatch] GetAccountNote failed", err);
      editingNote = "";
    }
  }

  function closeEditor(): void {
    editingRow = null;
  }

  async function handleSaveAccount(roles: OverwatchRoleRanks, note: string): Promise<void> {
    const target = editingRow;
    if (!target) return;
    const payload: Partial<Record<Role, RoleRankDTO>> = {};
    for (const role of ROLES) {
      const rr = roles[role];
      payload[role] = { tier: rr?.tier ?? "", division: rr?.division ?? 0 };
    }
    try {
      await Promise.all([
        OverwatchService.SetRanks(target.platformKey, target.id, payload),
        BasicService.SetAccountNote(target.platformKey, target.id, note),
      ]);
      rows = rows.map((r) => (rowKey(r) === rowKey(target) ? { ...r, roles } : r));
      editingRow = null;
    } catch (err) {
      pushToast({ type: "error", message: formatToastWithError("Could not save", err) });
    }
  }

  async function handleToggleHidden(): Promise<void> {
    const target = editingRow;
    if (!target) return;
    const nextHidden = !target.hidden;
    try {
      await OverwatchService.SetHidden(target.platformKey, target.id, nextHidden);
      rows = rows.map((r) => (rowKey(r) === rowKey(target) ? { ...r, hidden: nextHidden } : r));
      pushToast({
        type: "success",
        message: nextHidden ? `Hid ${target.name}. Turn on "Show hidden" to bring it back.` : `Unhid ${target.name}.`,
        duration: 4000,
      });
      editingRow = null;
    } catch (err) {
      pushToast({ type: "error", message: formatToastWithError("Could not update visibility", err) });
    }
  }

  $: visibleRows = showHidden ? rows : rows.filter((r) => !r.hidden);

  $: sortedRows = [...visibleRows].sort((a, b) => {
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
  <div class="ow-accounts-toolbar">
    <div class="ow-add-group">
      <img class="ow-add-icon" src={PLATFORM_ICONS.battlenet} alt="" />
      <button type="button" class="ow-add-btn" on:click={handleAddBattleNetAccount}>Add account</button>
      <button type="button" class="ow-add-btn" on:click={openSaveBattleNetPrompt}>Save current session</button>
    </div>
    <div class="ow-add-group">
      <img class="ow-add-icon" src={PLATFORM_ICONS.steam} alt="" />
      <button type="button" class="ow-add-btn" on:click={handleAddSteamAccount}>Add account</button>
      <button type="button" class="ow-add-btn" on:click={openSteamFolderPrompt}>Steam folder&hellip;</button>
    </div>
    <button type="button" class="ow-add-btn ow-add-btn--refresh" on:click={loadAll} disabled={refreshing}>
      {refreshing ? "Refreshing…" : "Refresh"}
    </button>
  </div>

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
    <label class="ow-show-hidden">
      <input type="checkbox" bind:checked={showHidden} />
      Show hidden
    </label>
  </div>

  <div class="ow-list">
    {#if initialLoading}
      <p class="ow-status">Loading accounts...</p>
    {:else if loadError}
      <p class="ow-status ow-status--error">{loadError}</p>
    {:else if sortedRows.length === 0 && rows.length > 0}
      <p class="ow-status">Every account is hidden. Turn on "Show hidden" above to bring one back.</p>
    {:else if sortedRows.length === 0}
      <p class="ow-status">No Battle.net or Steam accounts found yet. Sign in once from within each app and it will show up here.</p>
    {:else}
      {#each sortedRows as row (rowKey(row))}
        <OverwatchAccountCard {row} onLogin={() => handleLogin(row)} onEdit={() => openEditor(row)} />
      {/each}
    {/if}
  </div>
</div>

{#if editingRow}
  <OverwatchAccountEditor
    row={editingRow}
    note={editingNote}
    onSave={handleSaveAccount}
    onCancel={closeEditor}
    onToggleHidden={handleToggleHidden}
  />
{/if}

{#if showSaveBattleNetPrompt}
  <OverwatchNamePrompt
    title="Save current Battle.net session"
    body="This saves whichever Battle.net account is currently logged in as a new entry in the list below."
    initialValue={saveBattleNetSuggestedName}
    confirmLabel="Save"
    onConfirm={handleConfirmSaveBattleNet}
    onCancel={() => (showSaveBattleNetPrompt = false)}
  />
{/if}

{#if showSteamFolderPrompt}
  <OverwatchNamePrompt
    title="Steam install folder"
    body="The folder Steam is installed in, e.g. S:\Games\Steam. Leave as the default if Steam is installed normally."
    initialValue={steamFolderPromptValue}
    confirmLabel="Save"
    onConfirm={handleConfirmSteamFolder}
    onCancel={() => (showSteamFolderPrompt = false)}
  />
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
  .ow-accounts-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.75rem;
    flex: 0 0 auto;
    padding-bottom: 0.6rem;
    border-bottom: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
  }
  .ow-add-group {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }
  .ow-add-icon {
    width: 18px;
    height: 18px;
    border-radius: 50%;
  }
  .ow-add-btn {
    padding: 0.3rem 0.6rem;
    border-radius: 6px;
    border: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
    background: transparent;
    color: var(--whiteSecondary, #fff);
    font-size: 0.78rem;
    cursor: pointer;
  }
  .ow-add-btn:hover {
    background: var(--overlay-white-08, rgba(255, 255, 255, 0.08));
    border-color: var(--accent);
  }
  .ow-add-btn--refresh {
    margin-left: auto;
  }
  .ow-add-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  .ow-show-hidden {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    margin-left: auto;
    font-size: 0.78rem;
    color: var(--text-body-muted, #9d9d9d);
    cursor: pointer;
  }

  .ow-list {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 1rem;
    align-content: start;
    padding: 0.2rem 0.2rem 1rem;
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
