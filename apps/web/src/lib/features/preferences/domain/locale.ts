export type Locale = 'kk'|'ru';
export function localeFrom(value:unknown):Locale { return value==='ru'?'ru':'kk'; }
export const localeCookie='qmanga_locale';
export const localeKey='qmanga.locale';
export function languageTag(locale:Locale) { return locale==='ru'?'ru-RU':'kk-KZ'; }

