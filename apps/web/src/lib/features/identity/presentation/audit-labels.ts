import type {Locale} from '$lib/features/preferences/domain/locale';

export function coverAuditCopy(locale:Locale) {
 return locale==='ru'?{select_cover:'Выбор обложки',remove_cover:'Удаление обложки',field:'Обложка',empty:'Нет обложки'}:{select_cover:'Мұқабаны таңдау',remove_cover:'Мұқабаны алып тастау',field:'Мұқаба',empty:'Мұқаба жоқ'};
}

export function auditActorLabel(name:string,id:string,locale:Locale):string {
 return name.trim()||(id?'Аккаунт · '+id.slice(-8):locale==='ru'?'Аккаунт недоступен':'Аккаунт қолжетімсіз');
}

export function auditCoverValue(value:unknown,locale:Locale):string {
 const labels=coverAuditCopy(locale);
 return typeof value==='string'?(value?labels.field+' · '+value.slice(-8):labels.empty):'—';
}
