import type { Theme } from '../domain/theme';
import { themeFrom } from '../domain/theme';
export interface ThemeStore { read(): string | null; save(theme: Theme): void; apply(theme: Theme): void;readMotion?():boolean;saveMotion?(value:boolean):void;applyMotion?(value:boolean):void }
export function initializeTheme(store: ThemeStore): Theme { const theme = themeFrom(store.read()); store.apply(theme);store.applyMotion?.(store.readMotion?.()||false); return theme; }
export function changeTheme(store: ThemeStore, value: string): Theme { const theme = themeFrom(value); store.save(theme); store.apply(theme); return theme; }

export function changeMotion(store:ThemeStore,value:boolean):boolean{store.saveMotion?.(value);store.applyMotion?.(value);return value;}
