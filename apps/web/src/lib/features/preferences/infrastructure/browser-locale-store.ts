import type {LocaleStore} from '../application/locale';
import {localeFrom,localeKey,localeCookie,type Locale} from '../domain/locale';
export class BrowserLocaleStore implements LocaleStore {
 read():Locale { try { return localeFrom(localStorage.getItem(localeKey)); } catch { return 'kk'; } }
 save(locale:Locale) { try { localStorage.setItem(localeKey,locale); } catch {} document.cookie=localeCookie+'='+locale+'; Path=/; Max-Age=31536000; SameSite=Lax'+(location.protocol==='https:'?'; Secure':''); }
 apply(locale:Locale) { document.documentElement.lang=locale; }
}

