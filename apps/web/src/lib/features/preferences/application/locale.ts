import {localeFrom,type Locale} from '../domain/locale';
export interface LocaleStore { read():Locale; save(locale:Locale):void; apply(locale:Locale):void }
export function changeLocale(store:LocaleStore,value:unknown):Locale { const locale=localeFrom(value); store.save(locale); store.apply(locale); return locale; }

