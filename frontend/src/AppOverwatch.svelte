<script lang="ts">
  import { onMount } from "svelte";
  import TitleBar from "./components/TitleBar.svelte";
  import Toast from "./components/Toast.svelte";
  import OverwatchAccountList from "./pages/overwatch/OverwatchAccountList.svelte";
  import { appBarTitle } from "./stores/nav";

  onMount(() => {
    appBarTitle.set("Bonbon's Account Switcher");
  });
</script>

<div class="container">
  <TitleBar />
  <main class="content" id="app-main" tabindex="-1">
    <OverwatchAccountList />
  </main>
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
</style>
