import type {Locale} from '$lib/features/preferences/domain/locale';
export function catalogNames(locale:Locale) { return locale==='ru'?{
 kind:{manga:'Манга',manhwa:'Манхва',manhua:'Маньхуа'},
 original:{ongoing:'Выходит',completed:'Завершён',hiatus:'Перерыв',cancelled:'Отменён'},
 translation:{ongoing:'Переводится',completed:'Переведён',paused:'Перерыв',stopped:'Заброшен'}
}:{
 kind:{manga:'Манга',manhwa:'Манхва',manhua:'Маньхуа'},
 original:{ongoing:'Шығып жатыр',completed:'Аяқталған',hiatus:'Үзілісте',cancelled:'Тоқтатылған'},
 translation:{ongoing:'Аударылып жатыр',completed:'Аударылған',paused:'Үзілісте',stopped:'Тоқтатылған'}
};}
