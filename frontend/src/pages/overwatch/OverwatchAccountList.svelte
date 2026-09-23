<script lang="ts">
  import { onMount } from "svelte";
  import { Browser } from "@wailsio/runtime";
  import * as PlatformService from "../../../bindings/TcNo-Acc-Switcher/internal/platform/platformservice.js";
  import * as BasicService from "../../../bindings/TcNo-Acc-Switcher/internal/basic/basicservice.js";
  import * as SteamService from "../../../bindings/TcNo-Acc-Switcher/internal/steam/steamservice.js";
  import type { Settings as SteamSettings } from "../../../bindings/TcNo-Acc-Switcher/internal/steam/models.js";
  import * as OverwatchService from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/service.js";
  import type { RoleRankDTO } from "../../../bindings/TcNo-Acc-Switcher/internal/owrank/models.js";
  import * as OverwatchUpdateService from "../../../bindings/TcNo-Acc-Switcher/internal/owupdate/service.js";
  import { pushToast } from "../../stores/toast";
  import { formatToastWithError } from "../../lib/formatWailsError";
  import { PLATFORM_ICONS } from "../../lib/overwatch/icons";
  import { ROLES, Role, ROLE_LABELS, rankScore } from "../../lib/overwatch/rankLadder";
  import { OVERWATCH_THEMES, loadOverwatchTheme, applyOverwatchTheme, type OverwatchThemeId } from "../../lib/overwatch/theme";
  import OverwatchAccountCard from "./OverwatchAccountCard.svelte";
  import OverwatchAccountEditor from "./OverwatchAccountEditor.svelte";
  import OverwatchNamePrompt from "./OverwatchNamePrompt.svelte";
  import OverwatchUpdateDialog from "./OverwatchUpdateDialog.svelte";
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
  let showSaveBattleNetPrompt = false;
  let saveBattleNetSuggestedName = "";

  const DEFAULT_STEAM_FOLDER = "C:\\Program Files (x86)\\Steam";
  let showSteamFolderPrompt = false;
  let steamFolderPromptValue = DEFAULT_STEAM_FOLDER;
  let cachedSteamSettings: SteamSettings | null = null;

  let runOnStartup = false;
  let runOnStartupBusy = false;

  let currentTheme: OverwatchThemeId = loadOverwatchTheme();

  function handleSelectTheme(id: OverwatchThemeId): void {
    currentTheme = id;
    applyOverwatchTheme(id);
  }

  let checkingForUpdate = false;
  let availableUpdate: { version: string; notes: string; url: string } | null = null;

  // isManual distinguishes the toolbar button (always reports back, even
  // "you're up to date") from the launch-time background check (silent
  // unless it actually finds something - nobody wants a toast on every
  // startup just to confirm nothing changed).
  async function checkForUpdate(isManual: boolean): Promise<void> {
    if (checkingForUpdate) return;
    checkingForUpdate = true;
    try {
      const info = await OverwatchUpdateService.CheckForUpdate();
      if (info.available) {
        availableUpdate = { version: info.version, notes: info.notes, url: info.url };
      } else if (isManual) {
        pushToast({ type: "success", message: "You're up to date.", duration: 4000 });
      }
    } catch (err) {
      console.error("[overwatch] CheckForUpdate failed", err);
      if (isManual) {
        pushToast({ type: "error", message: formatToastWithError("Could not check for updates", err) });
      }
    } finally {
      checkingForUpdate = false;
    }
  }

  // Opens the release page in the user's own browser rather than downloading
  // and replacing this exe automatically - a self-updating binary is exactly
  // the behavior pattern antivirus heuristics flag as dropper/trojan-like,
  // which is what got an earlier version of this feature flagged.
  async function handleOpenDownloadPage(): Promise<void> {
    const url = availableUpdate?.url;
    availableUpdate = null;
    if (!url) return;
    try {
      await Browser.OpenURL(url);
    } catch (err) {
      console.error("[overwatch] OpenURL failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not open the download page", err) });
    }
  }

  async function handleToggleRunOnStartup(): Promise<void> {
    const next = !runOnStartup;
    runOnStartupBusy = true;
    try {
      await PlatformService.SetStartTrayWithWindows(next);
      runOnStartup = next;
    } catch (err) {
      console.error("[overwatch] SetStartTrayWithWindows failed", err);
      pushToast({ type: "error", message: formatToastWithError("Could not update startup setting", err) });
    } finally {
      runOnStartupBusy = false;
    }
  }

  function rowKey(row: { platform: string; id: string }): string {
    return `${row.platform}|${row.id}`;
  }

  async function loadMetaFor(
    platformKey: string,
    uniqueId: string,
  ): Promise<{ roles: OverwatchRoleRanks; hidden: boolean; favorite: boolean }> {
    try {
      const entry = await OverwatchService.GetRank(platformKey, uniqueId);
      const roles: OverwatchRoleRanks = {};
      for (const role of ROLES) {
        const rr = entry.roles[role];
        if (rr && rr.tier) roles[role] = { tier: rr.tier, division: rr.division };
      }
      return { roles, hidden: entry.hidden, favorite: entry.favorite };
    } catch {
      // An account with no ranks saved yet, or a store read failure - either
      // way the row just shows every role as unranked, visible, and unpinned.
      return { roles: {}, hidden: false, favorite: false };
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
            note: a.note,
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
            note: a.note,
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
    PlatformService.GetStartTrayWithWindows()
      .then((enabled) => { runOnStartup = enabled; })
      .catch((err) => console.error("[overwatch] GetStartTrayWithWindows failed", err));
    // Silent unless it actually finds something - see checkForUpdate.
    void checkForUpdate(false);
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

  function openEditor(row: OverwatchAccountRowData): void {
    editingRow = row;
  }

  function closeEditor(): void {
    editingRow = null;
  }

  async function handleSaveAccount(roles: OverwatchRoleRanks): Promise<void> {
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

  // Saved the moment the note field loses focus, independent of the ranks
  // Save button - typing a note then just closing the dialog (or the whole
  // app) used to lose it silently, since nothing had actually sent it to the
  // backend yet.
  async function handleSaveNote(note: string): Promise<void> {
    const target = editingRow;
    if (!target || note === target.note) return;
    try {
      await BasicService.SetAccountNote(target.platformKey, target.id, note);
      rows = rows.map((r) => (rowKey(r) === rowKey(target) ? { ...r, note } : r));
      if (editingRow && rowKey(editingRow) === rowKey(target)) {
        editingRow = { ...editingRow, note };
      }
    } catch (err) {
      pushToast({ type: "error", message: formatToastWithError("Could not save note", err) });
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

  async function handleToggleFavorite(row: OverwatchAccountRowData): Promise<void> {
    const nextFavorite = !row.favorite;
    try {
      await OverwatchService.SetFavorite(row.platformKey, row.id, nextFavorite);
      rows = rows.map((r) => (rowKey(r) === rowKey(row) ? { ...r, favorite: nextFavorite } : r));
    } catch (err) {
      pushToast({ type: "error", message: formatToastWithError("Could not update favorite", err) });
    }
  }

  $: visibleRows = showHidden ? rows : rows.filter((r) => !r.hidden);

  $: sortedRows = [...visibleRows].sort((a, b) => {
    // Favorites always float above every non-favorite account, regardless of
    // which sort is selected - within each of those two groups, the chosen
    // sort still applies.
    if (a.favorite !== b.favorite) return a.favorite ? -1 : 1;
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
    <label class="ow-run-on-startup">
      <input
        type="checkbox"
        checked={runOnStartup}
        disabled={runOnStartupBusy}
        on:change={handleToggleRunOnStartup}
      />
      Run on startup
    </label>
    <button type="button" class="ow-add-btn" on:click={() => checkForUpdate(true)} disabled={checkingForUpdate}>
      {checkingForUpdate ? "Checking…" : "Check for updates"}
    </button>
  </div>

  <div class="ow-toolbar">
    <span class="ow-toolbar-label">Theme</span>
    <div class="ow-theme-group" role="group" aria-label="Choose a theme">
      {#each OVERWATCH_THEMES as theme (theme.id)}
        <button
          type="button"
          class="ow-theme-swatch"
          class:ow-theme-swatch--active={currentTheme === theme.id}
          style="background: {theme.swatch};"
          title={theme.label}
          aria-label={theme.label}
          aria-pressed={currentTheme === theme.id}
          on:click={() => handleSelectTheme(theme.id)}
        ></button>
      {/each}
    </div>
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
        <OverwatchAccountCard
          {row}
          onLogin={() => handleLogin(row)}
          onEdit={() => openEditor(row)}
          onToggleFavorite={() => handleToggleFavorite(row)}
        />
      {/each}
    {/if}
  </div>
</div>

{#if editingRow}
  <OverwatchAccountEditor
    row={editingRow}
    onSave={handleSaveAccount}
    onSaveNote={handleSaveNote}
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

{#if availableUpdate}
  <OverwatchUpdateDialog
    version={availableUpdate.version}
    notes={availableUpdate.notes}
    onOpenDownload={handleOpenDownloadPage}
    onLater={() => (availableUpdate = null)}
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

  .ow-run-on-startup {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.78rem;
    color: var(--text-body-muted, #9d9d9d);
    cursor: pointer;
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

  .ow-theme-group {
    display: flex;
    gap: 0.5rem;
  }
  .ow-theme-swatch {
    width: 22px;
    height: 22px;
    border-radius: 50%;
    border: 2px solid transparent;
    padding: 0;
    cursor: pointer;
    box-shadow: 0 0 0 1px var(--overlay-white-14, rgba(255, 255, 255, 0.14));
  }
  .ow-theme-swatch:hover {
    box-shadow: 0 0 0 1px var(--accent);
  }
  .ow-theme-swatch--active {
    border-color: var(--mainContentBackground, #14181f);
    box-shadow: 0 0 0 2px var(--accent);
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
    color: var(--ow-accent-text, #0b0e12);
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
