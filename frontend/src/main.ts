import { mount } from 'svelte'
import { initI18n } from './stores/i18n'

const isOverwatchBuild = import.meta.env.VITE_APP_VARIANT === 'overwatch'

/**
 * Nothing before mount() may decide whether the app paints: the window is
 * frameless, so an app that never mounts is a bare rectangle with no title bar
 * and no way to report why. Each step degrades to its defaults instead.
 */
const STEP_TIMEOUT_MS = 8000

function guard(name: string, run: () => void): void {
  try {
    run()
  } catch (err) {
    console.error(`[boot] ${name} failed`, err)
  }
}

async function step(name: string, run: () => Promise<unknown>): Promise<void> {
  let timer: ReturnType<typeof setTimeout> | undefined
  try {
    await Promise.race([
      run(),
      new Promise((_, reject) => {
        timer = setTimeout(
          () => reject(new Error(`${name} timed out after ${STEP_TIMEOUT_MS}ms`)),
          STEP_TIMEOUT_MS,
        )
      }),
    ])
  } catch (err) {
    console.error(`[boot] ${name} failed`, err)
  } finally {
    clearTimeout(timer)
  }
}

function mountInto(component: any, target: HTMLElement): void {
  try {
    mount(component, { target })
    window.__tcnoBoot?.ready()
  } catch (err) {
    window.__tcnoBoot?.fail('mount', err)
    throw err
  }
}

async function bootOverwatch(): Promise<void> {
  await import('./styles/normalize.scss')
  // theme.scss supplies every CSS custom property the reused components
  // (TitleBar, AccountLiveSessionIndicator, Toast) expect, and style.scss the
  // base font/scrollbar/window-sizing rules; overwatch.scss layers this
  // build's own small set of named themes on top, applied below before mount
  // so the first paint is never the wrong one.
  await import('./styles/theme.scss')
  await import('./styles/style.scss')
  await import('./styles/overwatch.scss')
  const { loadOverwatchTheme, applyOverwatchTheme } = await import('./lib/overwatch/theme')
  applyOverwatchTheme(loadOverwatchTheme())
  // TitleBar (reused as-is for window drag/close) reads translated strings, so
  // i18n still loads; everything else the full app boots - offline mode,
  // routing, the navigation guard - has no equivalent here since there is
  // exactly one screen.
  await step('i18n', initI18n)
  const { default: AppOverwatch } = await import('./AppOverwatch.svelte')
  const target = document.getElementById('app')
  if (!target) {
    throw new Error('#app is missing from the document')
  }
  mountInto(AppOverwatch, target)
}

async function bootMain(): Promise<void> {
  await import('./styles/context_menu.scss')
  await import('./styles/normalize.scss')
  await import('./styles/style.scss')
  await import('./styles/theme.scss')
  await import('./styles/overlayReceivers.scss')
  await import('./styles/UI.scss')
  await import('./styles/modal-primary.scss')
  await import('./styles/acclist.scss')
  await import('./styles/rtl.scss')
  const { initOfflineMode } = await import('./stores/offlineMode')
  const { resolveInitialRoute, installHashSync } = await import('./stores/nav')
  const { initTheme } = await import('./lib/themes')
  const { installNavigationGuard } = await import('./lib/navigationGuard')
  const { default: App } = await import('./App.svelte')

  guard('navigation guard', installNavigationGuard)

  await step('i18n', initI18n)
  await step('offline mode', initOfflineMode)
  await step('theme', initTheme)
  await step('initial route', resolveInitialRoute)
  guard('hash sync', installHashSync)

  const target = document.getElementById('app')
  if (!target) {
    throw new Error('#app is missing from the document')
  }
  mountInto(App, target)
}

void (isOverwatchBuild ? bootOverwatch() : bootMain())
