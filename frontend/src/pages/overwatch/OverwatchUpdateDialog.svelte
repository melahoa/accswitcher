<script lang="ts">
  export let version: string;
  export let notes: string;
  export let onOpenDownload: () => void;
  export let onLater: () => void;

  function handleBackdropKeydown(e: KeyboardEvent): void {
    if (e.key === "Escape") onLater();
  }
</script>

<svelte:window on:keydown={handleBackdropKeydown} />

<div
  class="ow-update-backdrop"
  role="presentation"
  on:click={(e) => { if (e.target === e.currentTarget) onLater(); }}
>
  <div class="ow-update" role="dialog" tabindex="-1" aria-modal="true" aria-label="Update available">
    <h2 class="ow-update-title">Update available: v{version}</h2>
    {#if notes}
      <p class="ow-update-notes">{notes}</p>
    {/if}
    <p class="ow-update-hint">This opens the release page in your browser - the download and install are up to you.</p>
    <div class="ow-update-actions">
      <button type="button" class="ow-btn ow-btn--secondary" on:click={onLater}>Later</button>
      <button type="button" class="ow-btn ow-btn--primary" on:click={onOpenDownload}>Open download page</button>
    </div>
  </div>
</div>

<style>
  .ow-update-backdrop {
    position: fixed;
    inset: 0;
    z-index: 110;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--backdrop-scrim-55, rgba(0, 0, 0, 0.55));
  }
  .ow-update {
    width: min(400px, calc(100vw - 2rem));
    padding: 1.25rem;
    border-radius: 10px;
    background: var(--mainContentBackground, #14181f);
    box-shadow: var(--shadow, 0 4px 10px -6px #010101);
    color: var(--whiteSecondary, #fff);
  }
  .ow-update-title {
    margin: 0 0 0.5rem;
    font-size: 1rem;
  }
  .ow-update-notes {
    margin: 0 0 1rem;
    font-size: 0.85rem;
    color: var(--text-body-muted, #9d9d9d);
    white-space: pre-line;
    max-height: 220px;
    overflow-y: auto;
  }
  .ow-update-hint {
    margin: 0 0 1rem;
    font-size: 0.78rem;
    color: var(--text-dim-gray, #6b6a6a);
  }
  .ow-update-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
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
    color: var(--ow-accent-text, #0b0e12);
    font-weight: 600;
  }
  .ow-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
