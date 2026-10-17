import type {Locale} from '$lib/features/preferences/domain/locale';
import {text} from './locale';
const errors={
 request_limit:['Өтінімдер шегіне жеттің. Алдыңғылары қаралғанша күт.','Достигнут лимит заявок. Дождись рассмотрения предыдущих.'],
 network:['Желіге қосылып, қайта көр.','Подключись к сети и попробуй снова.'],
 unavailable:['Қызмет уақытша қолжетімсіз. Кейінірек қайтала.','Сервис временно недоступен. Попробуй позже.'],
 credentials:['Пошта немесе құпиясөз дұрыс емес.','Неверная почта или пароль.'],
 unauthorized:['Аккаунтқа қайта кір.','Войди в аккаунт снова.'],
 forbidden:['Бұл әрекетке рұқсат жоқ.','Нет доступа к этому действию.'],
 conflict:['Деректер өзгерген. Бетті жаңартып, қайта көр.','Данные изменились. Обнови страницу и повтори.'],
 rate_limit:['Сұрау тым көп. Сәл күтіп, қайтала.','Слишком много запросов. Подожди немного.'],
 validation:['Енгізілген деректерді тексер.','Проверь введённые данные.'],
 email:['Пошта мекенжайын тексер.','Проверь адрес почты.'],
 password:['Құпиясөз 10–128 таңбадан тұруы керек.','Пароль должен содержать 10–128 символов.'],
 expired:['Сілтеменің мерзімі өткен. Жаңасын сұрат.','Ссылка истекла. Запроси новую.'],
 not_found:['Табылмады немесе енді қолжетімсіз.','Не найдено или больше недоступно.'],
 csrf:['Бетті жаңартып, қайта көр.','Обнови страницу и повтори.'],
 origin:['Бетті жаңартып, қайта көр.','Обнови страницу и повтори.']
} satisfies Record<string,readonly [string,string]>;
export function errorText(locale:Locale,code:unknown):string {
 const pair=typeof code==='string' && Object.hasOwn(errors,code)?errors[code as keyof typeof errors]:errors.unavailable;
 return text(locale,pair[0],pair[1]);
}
export function failureText(locale:Locale,failure:unknown):string {
 return errorText(locale,typeof failure==='object'&&failure!==null&&'code' in failure?failure.code:'unavailable');
}
