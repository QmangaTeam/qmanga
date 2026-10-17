import type { Theme } from '../domain/theme';
import type { ThemeStore } from '../application/theme';
export class BrowserThemeStore implements ThemeStore {
 read() { try { return localStorage.getItem('qmanga.theme'); } catch { return null; } }
 save(theme: Theme) { try { localStorage.setItem('qmanga.theme', theme); } catch {} }
 apply(theme: Theme) { if (theme === 'system') delete document.documentElement.dataset.theme; else document.documentElement.dataset.theme = theme; }
}
