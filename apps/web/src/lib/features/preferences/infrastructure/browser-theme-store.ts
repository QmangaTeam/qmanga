import type { Theme } from '../domain/theme';
import type { ThemeStore } from '../application/theme';
export class BrowserThemeStore implements ThemeStore {
 readMotion(){try{return localStorage.getItem('qmanga.motion-reduced')==='true';}catch{return false;}}
 saveMotion(value:boolean){try{localStorage.setItem('qmanga.motion-reduced',String(value));}catch{}}
 applyMotion(value:boolean){if(value)document.documentElement.dataset.motion='reduce';else delete document.documentElement.dataset.motion;}
 read() { try { return localStorage.getItem('qmanga.theme'); } catch { return null; } }
 save(theme: Theme) { try { localStorage.setItem('qmanga.theme', theme); } catch {} }
 apply(theme: Theme) { if (theme === 'system') delete document.documentElement.dataset.theme; else document.documentElement.dataset.theme = theme; }
}
