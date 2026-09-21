export type OverwatchThemeId = "dark" | "light" | "cherry" | "crimson" | "ocean";

export interface OverwatchThemeDef {
  id: OverwatchThemeId;
  label: string;
  /** CSS background value for the picker's preview swatch. */
  swatch: string;
}

export const OVERWATCH_THEMES: OverwatchThemeDef[] = [
  { id: "dark", label: "Default", swatch: "linear-gradient(135deg, #3a3a3a, #0c0c0c)" },
  { id: "light", label: "Light", swatch: "linear-gradient(135deg, #ffffff, #c7d2e0)" },
  { id: "cherry", label: "Cherry Blossom", swatch: "linear-gradient(135deg, #ffffff, #ff8fc4)" },
  { id: "crimson", label: "Crimson", swatch: "linear-gradient(135deg, #e0293f, #0a0000)" },
  { id: "ocean", label: "Ocean", swatch: "linear-gradient(135deg, #cdeeff, #0b2545)" },
];

const STORAGE_KEY = "overwatch-theme";
const DEFAULT_THEME: OverwatchThemeId = "dark";

function isThemeId(value: string | null): value is OverwatchThemeId {
  return OVERWATCH_THEMES.some((t) => t.id === value);
}

/** Reads the saved theme choice. Never throws - a private window or blocked storage just falls back to the default. */
export function loadOverwatchTheme(): OverwatchThemeId {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (isThemeId(stored)) return stored;
  } catch {
    /* storage unavailable - use the default */
  }
  return DEFAULT_THEME;
}

/** Applies the theme to the document and remembers the choice for next launch. */
export function applyOverwatchTheme(theme: OverwatchThemeId): void {
  document.documentElement.setAttribute("data-ow-theme", theme);
  try {
    localStorage.setItem(STORAGE_KEY, theme);
  } catch {
    /* per-viewer convenience only - fine if it doesn't persist */
  }
}
