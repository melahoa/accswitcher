<script lang="ts">
  export let title: string;
  export let body: string;
  export let initialValue: string = "";
  export let confirmLabel: string = "Save";
  export let onConfirm: (value: string) => void;
  export let onCancel: () => void;

  let value = initialValue;
  let inputEl: HTMLInputElement | undefined;

  function focusInput(el: HTMLInputElement): void {
    inputEl = el;
    el.focus();
    el.select();
  }

  function submit(): void {
    const trimmed = value.trim();
    if (!trimmed) return;
    onConfirm(trimmed);
  }

  function handleKeydown(e: KeyboardEvent): void {
    if (e.key === "Escape") onCancel();
  }

  function handleInputKeydown(e: KeyboardEvent): void {
    if (e.key === "Enter") {
      e.preventDefault();
      submit();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div
  class="ow-prompt-backdrop"
  role="presentation"
  on:click={(e) => { if (e.target === e.currentTarget) onCancel(); }}
>
  <div class="ow-prompt" role="dialog" tabindex="-1" aria-modal="true" aria-label={title}>
    <h2 class="ow-prompt-title">{title}</h2>
    <p class="ow-prompt-body">{body}</p>
    <input
      class="ow-prompt-input"
      type="text"
      bind:value
      use:focusInput
      on:keydown={handleInputKeydown}
    />
    <div class="ow-prompt-actions">
      <button type="button" class="ow-btn ow-btn--secondary" on:click={onCancel}>Cancel</button>
      <button type="button" class="ow-btn ow-btn--primary" disabled={!value.trim()} on:click={submit}>{confirmLabel}</button>
    </div>
  </div>
</div>

<style>
  .ow-prompt-backdrop {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--backdrop-scrim-55, rgba(0, 0, 0, 0.55));
  }
  .ow-prompt {
    width: min(380px, calc(100vw - 2rem));
    padding: 1.25rem;
    border-radius: 10px;
    background: var(--mainContentBackground, #14181f);
    box-shadow: var(--shadow, 0 4px 10px -6px #010101);
    color: var(--whiteSecondary, #fff);
  }
  .ow-prompt-title {
    margin: 0 0 0.5rem;
    font-size: 1rem;
  }
  .ow-prompt-body {
    margin: 0 0 0.75rem;
    font-size: 0.85rem;
    color: var(--text-body-muted, #9d9d9d);
  }
  .ow-prompt-input {
    width: 100%;
    padding: 0.5rem 0.6rem;
    border-radius: 6px;
    background: var(--code-background, #131a20);
    color: var(--whiteSecondary, #fff);
    border: 1px solid var(--overlay-white-14, rgba(255, 255, 255, 0.14));
    font-size: 0.9rem;
  }
  .ow-prompt-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    margin-top: 1rem;
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
  .ow-btn--primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
