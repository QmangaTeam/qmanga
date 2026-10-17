import {test,expect} from 'bun:test';
import {localeFrom} from './locale';
import {changeLocale,type LocaleStore} from '../application/locale';
test('only an explicit Russian choice overrides Kazakh',()=>{
 for(const value of [null,undefined,'','RU','ru-RU','en','kk']) expect(localeFrom(value)).toBe('kk');
 expect(localeFrom('ru')).toBe('ru');
});
test('locale changes save and apply without touching reading state',()=>{
 const events:string[]=[];
 const store:LocaleStore={read:()=> 'kk',save:value=>events.push('save:'+value),apply:value=>events.push('apply:'+value)};
 expect(changeLocale(store,'ru')).toBe('ru');
 expect(events).toEqual(['save:ru','apply:ru']);
});

