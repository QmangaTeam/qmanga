import type { Theme } from '../domain/theme';
import { themeFrom } from '../domain/theme';
export interface ThemeStore { read(): string | null; save(theme: Theme): void; apply(theme: Theme): void }
export function initializeTheme(store: ThemeStore): Theme { const theme = themeFrom(store.read()); store.apply(theme); return theme; }
export function changeTheme(store: ThemeStore, value: string): Theme { const theme = themeFrom(value); store.save(theme); store.apply(theme); return theme; }
