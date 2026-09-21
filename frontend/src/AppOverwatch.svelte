<script lang="ts">
  import { onMount } from "svelte";
  import TitleBar from "./components/TitleBar.svelte";
  import Toast from "./components/Toast.svelte";
  import OverwatchAccountList from "./pages/overwatch/OverwatchAccountList.svelte";
  import { appBarTitle } from "./stores/nav";
  import * as OverwatchUpdateService from "../bindings/TcNo-Acc-Switcher/internal/owupdate/service.js";

  let version = "";

  onMount(() => {
    appBarTitle.set("Bonbon's Account Switcher");
    OverwatchUpdateService.Version()
      .then((v) => { version = v; })
      .catch((err) => console.error("[overwatch] Version failed", err));
  });
</script>

<div class="container">
  <TitleBar />
  <main class="content" id="app-main" tabindex="-1">
    <OverwatchAccountList />
  </main>
  {#if version}
    <div class="version-badge">v{version}</div>
  {/if}
  <Toast />
</div>

<style>
  .container {
    background: var(--program-bg);
    height: 100vh;
    width: 100vw;
    display: flex;
    flex-direction: column;
  }
  .content {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-left: var(--border-bar-size) solid var(--border-bar-bg);
    border-right: var(--border-bar-size) solid var(--border-bar-bg);
    border-bottom: var(--border-bar-size) solid var(--border-bar-bg);
  }
  .version-badge {
    position: fixed;
    right: 8px;
    bottom: 4px;
    z-index: 5;
    font-size: 0.7rem;
    color: var(--text-dim-gray, #6b6a6a);
    pointer-events: none;
    user-select: none;
  }
</style>
