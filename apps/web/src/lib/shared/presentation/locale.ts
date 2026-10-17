import {getContext,setContext} from 'svelte';
import {writable,type Writable} from 'svelte/store';
import type {Locale} from '$lib/features/preferences/domain/locale';
export interface LocaleContext { locale:Writable<Locale>; change:(value:Locale)=>void }
const contextKey=Symbol('qmanga.locale');
export function provideLocale(initial:Locale,save:(value:Locale)=>void):LocaleContext {
 const locale=writable(initial);
 const context={locale,change:(value:Locale)=>{save(value);locale.set(value);}};
 setContext(contextKey,context);
 return context;
}
export function useLocale():LocaleContext {
 return getContext<LocaleContext>(contextKey) || {locale:writable<Locale>('kk'),change:()=>{}};
}
export function text(locale:Locale,kazakh:string,russian:string):string { return locale==='ru'?russian:kazakh; }

