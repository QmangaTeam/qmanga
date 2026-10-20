import type {Locale} from '../../features/preferences/domain/locale';
import './form-validation.css';
type Constraints=Partial<Pick<ValidityState,'valueMissing'|'tooShort'|'tooLong'|'typeMismatch'|'patternMismatch'|'rangeUnderflow'|'rangeOverflow'|'stepMismatch'|'badInput'>>;
export function constraintMessage(validity:Constraints,locale:Locale){
 const ru=locale==='ru';
 if(validity.valueMissing)return ru?'Заполни это поле.':'Осы өрісті толтыр.';
 if(validity.tooShort)return ru?'Текст слишком короткий.':'Мәтін тым қысқа.';
 if(validity.tooLong)return ru?'Текст слишком длинный.':'Мәтін тым ұзын.';
 if(validity.typeMismatch||validity.patternMismatch)return ru?'Проверь формат значения.':'Мәннің пішімін тексер.';
 if(validity.rangeUnderflow||validity.rangeOverflow||validity.stepMismatch||validity.badInput)return ru?'Укажи допустимое число.':'Жарамды санды көрсет.';
 return ru?'Проверь значение поля.':'Өріс мәнін тексер.';
}
type Field=HTMLInputElement|HTMLTextAreaElement|HTMLSelectElement;
let sequence=0;
/** Native constraints remain authoritative; only field errors and focus are enhanced. */
export function fieldValidation(form:HTMLFormElement,options:{locale:Locale}){
 const errors=new Map<Field,{node:HTMLSpanElement;described:string|null;invalid:string|null}>();
 const summary=document.createElement('div');summary.className='validation-summary';summary.setAttribute('role','alert');summary.tabIndex=-1;summary.hidden=true;form.prepend(summary);
 function reveal(field:Field){let parent=field.parentElement;while(parent){if(parent instanceof HTMLDetailsElement)parent.open=true;parent=parent.parentElement;}}
 function label(field:Field){const source=field.labels?.[0]?.cloneNode(true) as HTMLElement|undefined;source?.querySelectorAll('input,textarea,select,.field-error,small').forEach(node=>node.remove());return field.getAttribute('aria-label')||source?.textContent?.trim()||field.name||field.id;}
 function render(){summary.replaceChildren();summary.hidden=errors.size===0;if(!errors.size)return;const title=document.createElement('p');title.textContent=options.locale==='ru'?'Исправь поля перед отправкой:':'Жібермес бұрын өрістерді түзет:';summary.append(title);const list=document.createElement('ul');for(const [field,{node}]of errors){const item=document.createElement('li'),button=document.createElement('button');button.type='button';button.textContent=label(field)+': '+node.textContent;button.onclick=()=>{reveal(field);field.focus();};item.append(button);list.append(item);}summary.append(list);}
 function clear(field:Field){const item=errors.get(field);if(!item)return;item.node.remove();if(item.described===null)field.removeAttribute('aria-describedby');else field.setAttribute('aria-describedby',item.described);if(item.invalid===null)field.removeAttribute('aria-invalid');else field.setAttribute('aria-invalid',item.invalid);errors.delete(field);}
 function show(field:Field){let item=errors.get(field);if(!item){const node=document.createElement('span');node.className='field-error';node.id='field-error-'+ ++sequence;item={node,described:field.getAttribute('aria-describedby'),invalid:field.getAttribute('aria-invalid')};errors.set(field,item);(field.closest('label')||field).insertAdjacentElement('afterend',node);field.setAttribute('aria-describedby',[item.described,node.id].filter(Boolean).join(' '));field.setAttribute('aria-invalid','true');}item.node.textContent=constraintMessage(field.validity,options.locale);reveal(field);}
 let scheduled=false;
 const invalid=(event:Event)=>{const field=event.target;if(!(field instanceof HTMLInputElement||field instanceof HTMLTextAreaElement||field instanceof HTMLSelectElement))return;event.preventDefault();show(field);if(!scheduled){scheduled=true;queueMicrotask(()=>{scheduled=false;render();summary.focus();});}};
 const input=(event:Event)=>{const field=event.target as Field;if(!errors.has(field))return;if(field.validity.valid)clear(field);else show(field);render();};
 const reset=()=>{for(const field of errors.keys())clear(field);render();};
 form.addEventListener('invalid',invalid,true);form.addEventListener('input',input);form.addEventListener('change',input);form.addEventListener('reset',reset);
 return {update(next:{locale:Locale}){options=next;for(const field of errors.keys())show(field);render();},destroy(){form.removeEventListener('invalid',invalid,true);form.removeEventListener('input',input);form.removeEventListener('change',input);form.removeEventListener('reset',reset);reset();summary.remove();}};
}
