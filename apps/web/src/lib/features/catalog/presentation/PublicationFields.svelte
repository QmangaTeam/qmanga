<script lang="ts">
 import {emptyPublication,type Publication} from '../domain/publication';
 import {useLocale,text} from '$lib/shared/presentation/locale';
 let {value=$bindable<Publication>(emptyPublication())}:{value?:Publication}=$props();
 const {locale}=useLocale();
</script>
<fieldset><legend>{text($locale,'Басылым туралы','Сведения об издании')}</legend>
 <label>{text($locale,'Шыққан жылы','Год выпуска')}<input type="number" min="1000" max="2200" value={value.year||''} oninput={event=>value.year=Number(event.currentTarget.value)||0}/></label>
 <label>{text($locale,'Баспалар, әр жолға біреуі','Издательства, по одному на строке')}<textarea rows="2" maxlength="2000" value={value.publishers.join('\n')} onchange={event=>value.publishers=event.currentTarget.value.split('\n').map(v=>v.trim()).filter(Boolean)}></textarea></label>
 <label>{text($locale,'Түпнегіз','Первоисточник')}<input bind:value={value.source} maxlength="160"/></label>
 <label>{text($locale,'Шыққан елі','Страна производства')}<input bind:value={value.country} maxlength="160"/></label>
 <label>{text($locale,'Форматтар, әр жолға біреуі','Форматы, по одному на строке')}<textarea rows="2" maxlength="1400" value={value.formats.join('\n')} onchange={event=>value.formats=event.currentTarget.value.split('\n').map(v=>v.trim()).filter(Boolean)}></textarea></label>
 <h3>{text($locale,'Сыртқы сілтемелер','Внешние ссылки')}</h3>
 {#each value.links as link,index}<div class="link-fields"><label>{text($locale,'Сілтеме атауы','Название ссылки')}<input bind:value={link.label} maxlength="80" required/></label><label>URL<input type="url" bind:value={link.url} maxlength="2000" required/></label><button class="button subtle" type="button" onclick={()=>value.links=value.links.filter((_,i)=>i!==index)}>{text($locale,'Алып тастау','Убрать')}</button></div>{/each}
 <button class="button subtle" type="button" disabled={value.links.length>=12} onclick={()=>value.links=[...value.links,{label:'',url:''}]}>{text($locale,'Сілтеме қосу','Добавить ссылку')}</button>
</fieldset>
<style>fieldset{display:grid;gap:12px;border:1px solid var(--line);padding:12px;border-radius:8px;margin:16px 0}label{display:grid;gap:6px;font-size:13px}.link-fields{display:grid;gap:8px;padding:12px 0;border-top:1px solid var(--line)}h3{font-size:15px;margin:8px 0}button{justify-self:start}</style>
