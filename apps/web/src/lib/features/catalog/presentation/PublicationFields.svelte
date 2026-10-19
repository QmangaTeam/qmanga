<script lang="ts">
	import { emptyPublication, type Publication } from '../domain/publication';
	import { countryFacets, formatFacets, sourceFacets } from '../domain/facets';
	import { useLocale, text } from '$lib/shared/presentation/locale';
	let { value = $bindable<Publication>(emptyPublication()) }: { value?: Publication } = $props();
	const { locale } = useLocale();
	function toggleFormat(format: string) {
		value.formats = value.formats.includes(format) ? value.formats.filter(item => item !== format) : [...value.formats, format];
	}
</script>
<fieldset><legend>{text($locale,'Басылым туралы','Сведения об издании')}</legend>
 <label>{text($locale,'Шыққан жылы','Год выпуска')}<input type="number" min="1000" max="2200" value={value.year||''} oninput={event=>value.year=Number(event.currentTarget.value)||0}/></label>
 <label>{text($locale,'Баспалар, әр жолға біреуі','Издательства, по одному на строке')}<textarea rows="2" maxlength="2000" value={value.publishers.join('\n')} onchange={event=>value.publishers=event.currentTarget.value.split('\n').map(v=>v.trim()).filter(Boolean)}></textarea></label>
 <label>{text($locale,'Түпнегіз','Первоисточник')}<select bind:value={value.source}><option value="">{text($locale,'Көрсетілмеген','Не указано')}</option>{#each sourceFacets as item}<option value={item.value}>{$locale==='ru'?item.ru:item.kk}</option>{/each}</select></label>
 <label>{text($locale,'Шыққан елі','Страна производства')}<select bind:value={value.country}><option value="">{text($locale,'Көрсетілмеген','Не указано')}</option>{#each countryFacets as item}<option value={item.value}>{$locale==='ru'?item.ru:item.kk}</option>{/each}</select></label>
 <fieldset class="formats"><legend>{text($locale,'Форматтар','Форматы')}</legend>{#each formatFacets as item}<label class="format-option"><input type="checkbox" checked={value.formats.includes(item.value)} onchange={()=>toggleFormat(item.value)}/><span>{$locale==='ru'?item.ru:item.kk}</span></label>{/each}</fieldset>
 <h3>{text($locale,'Сыртқы сілтемелер','Внешние ссылки')}</h3>
 {#each value.links as link,index}<div class="link-fields"><label>{text($locale,'Сілтеме атауы','Название ссылки')}<input bind:value={link.label} maxlength="80" required/></label><label>URL<input type="url" bind:value={link.url} maxlength="2000" required/></label><button class="button subtle" type="button" onclick={()=>value.links=value.links.filter((_,i)=>i!==index)}>{text($locale,'Алып тастау','Убрать')}</button></div>{/each}
 <button class="button subtle" type="button" disabled={value.links.length>=12} onclick={()=>value.links=[...value.links,{label:'',url:''}]}>{text($locale,'Сілтеме қосу','Добавить ссылку')}</button>
</fieldset>
<style>fieldset{display:grid;gap:12px;border:1px solid var(--line);padding:12px;border-radius:8px;margin:16px 0}fieldset.formats{grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:4px 12px}label{display:grid;gap:6px;font-size:13px}.format-option{display:flex;align-items:center;gap:8px;min-height:44px;font-size:14px}.format-option input{width:20px;height:20px;margin:0;accent-color:var(--accent)}.link-fields{display:grid;gap:8px;padding:12px 0;border-top:1px solid var(--line)}h3{font-size:15px;margin:8px 0}button{justify-self:start}</style>
